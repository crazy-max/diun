package app

import (
	stderrors "errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/crazy-max/diun/v4/internal/matcher"
	"github.com/crazy-max/diun/v4/internal/model"
	"github.com/crazy-max/diun/v4/pkg/registry"
	"github.com/crazy-max/diun/v4/pkg/registry/auth"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	podmanmanifest "go.podman.io/image/v5/manifest"
)

// jobPlan is built on the discovery goroutine before any analysis writes to the
// database. Clients with identical effective options share concrete targets.
type jobPlan struct {
	clients map[registry.Options]*registry.Client
	groups  map[analysisKey]*analysisJob
	jobs    []*analysisJob
	tags    map[analysisKey]tagList
}

type tagList struct {
	tags []string
	err  error
}

type analysisKey struct {
	image  string
	client *registry.Client
}

type analysisJob struct {
	consumers []model.Job
	baseline  registry.Manifest
	err       error
}

func (di *Diun) createJob(job model.Job, plan *jobPlan) {
	var err error
	var prvImage registry.Image

	sublog := log.With().
		Str("provider", job.Provider).
		Str("image", job.Image.Name).
		Logger()

	// Parse image
	prvImage, err = registry.ParseImage(registry.ParseImageOptions{
		Name:   job.Image.Name,
		HubTpl: job.Image.HubTpl,
	})
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot parse image")
		return
	}
	job.RegImage = prvImage
	job.HubLinkOverride = job.Image.HubLink

	// First check?
	job.FirstCheck, err = di.db.First(job.RegImage)
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot check first")
		return
	}

	// Get registry options
	reg, err := di.cfg.RegOpts.Select(job.Image.RegOpt, job.RegImage)
	if err != nil {
		sublog.Warn().Err(err).Msg("Registry options")
		return
	} else if reg != nil {
		sublog.Debug().Str("regopt", reg.Name).Msg("Registry options will be used")
	} else {
		reg = (&model.RegOpt{}).GetDefaults()
	}

	// Validate include/exclude tags
	for _, includeTag := range job.Image.IncludeTags {
		if _, err := regexp.Compile(includeTag); err != nil {
			sublog.Error().Err(err).Msg("Include tag regex '%s' cannot compile")
			return
		}
	}
	for _, excludeTag := range job.Image.ExcludeTags {
		if _, err := regexp.Compile(excludeTag); err != nil {
			sublog.Error().Err(err).Msg("Exclude tag regex '%s' cannot compile")
			return
		}
	}

	authProvider, err := plan.loadAuth(reg, job.RegImage.Domain)
	if err != nil {
		sublog.Warn().Err(err).Str("regopt", reg.Name).Msg("Cannot load registry credentials")
		if authProvider == nil {
			return
		}
	}

	opts := registry.Options{
		AuthProvider:  authProvider,
		Timeout:       *reg.Timeout,
		InsecureTLS:   *reg.InsecureTLS,
		UserAgent:     di.meta.UserAgent,
		CompareDigest: *di.cfg.Watch.CompareDigest,
		ImageOs:       job.Image.Platform.OS,
		ImageArch:     job.Image.Platform.Arch,
		ImageVariant:  job.Image.Platform.Variant,
	}

	if plan.clients == nil {
		plan.clients = make(map[registry.Options]*registry.Client)
	}
	job.Registry = plan.clients[opts]
	if job.Registry == nil {
		job.Registry, err = registry.New(opts)
		if err != nil {
			sublog.Error().Err(err).Msg("Cannot create registry client")
			return
		}
		plan.clients[opts] = job.Registry
	}

	di.addJob(job, plan)

	if job.Image.WatchMode == registry.WatchModeTag || len(job.RegImage.Domain) == 0 {
		return
	}

	tags, err := plan.selectTags(job, registry.TagsOptions{
		WatchMode: job.Image.WatchMode,
		Image:     job.RegImage,
		Max:       job.Image.MaxTags,
		Sort:      job.Image.SortTags,
		Include:   job.Image.IncludeTags,
		Exclude:   job.Image.ExcludeTags,
	})
	if err != nil {
		sublog.Error().Err(err).Msg("Cannot list tags from registry")
		return
	}

	log.Debug().Str("image", job.RegImage.String()).Msgf("%d tag(s) found in repository. %d will be analyzed (%d max, %d not included, %d excluded, %d artifact tags, %d outside watch mode).",
		tags.Total,
		len(tags.List),
		job.Image.MaxTags,
		tags.NotIncluded,
		tags.Excluded,
		tags.Artifacts,
		tags.OutsideWatchMode,
	)

	for _, tag := range tags.List {
		if prvImage.Tag == tag {
			continue
		}
		job.Image.Name = fmt.Sprintf("%s/%s:%s", job.RegImage.Domain, job.RegImage.Path, tag)
		job.RegImage, err = registry.ParseImage(registry.ParseImageOptions{
			Name:   job.Image.Name,
			HubTpl: job.Image.HubTpl,
		})
		if err != nil {
			sublog.Error().Err(err).Msg("Cannot parse image (tag)")
			continue
		}
		di.addJob(job, plan)
	}
}

// loadAuth returns nil on configured provider setup failures. Docker discovery
// failures return an anonymous provider alongside the error for warn-and-continue.
func (plan *jobPlan) loadAuth(reg *model.RegOpt, domain string) (auth.Provider, error) {
	switch reg.Auth {
	case nil:
		// TODO: Remove legacy top-level credential support in a future release. Users should migrate to auth.static.
		provider, staticErr := auth.NewStatic(auth.StaticOptions{
			Username:     reg.Username,
			UsernameFile: reg.UsernameFile,
			Password:     reg.Password,
			PasswordFile: reg.PasswordFile,
		})
		if staticErr != nil {
			return nil, staticErr
		}
		if provider.Username != "" {
			return provider, nil
		}
		docker, err := auth.NewDocker(domain)
		return docker, errors.Wrap(err, "error seeking Docker credentials")
	default:
		provider, err := auth.NewStatic(*reg.Auth.Static)
		if err != nil {
			return nil, errors.Wrap(err, "cannot configure static credentials")
		}
		return provider, nil
	}
}

func (plan *jobPlan) selectTags(job model.Job, opts registry.TagsOptions) (*registry.Tags, error) {
	if _, err := opts.Select(nil); err != nil {
		return nil, err
	}
	if plan.tags == nil {
		plan.tags = make(map[analysisKey]tagList)
	}
	key := analysisKey{image: job.RegImage.Name(), client: job.Registry}
	list, ok := plan.tags[key]
	if !ok {
		list.tags, list.err = job.Registry.ListTags(job.RegImage)
		plan.tags[key] = list
	}
	if list.err != nil {
		return nil, list.err
	}
	return opts.Select(list.tags)
}

func (di *Diun) addJob(job model.Job, plan *jobPlan) {
	if plan.groups == nil {
		plan.groups = make(map[analysisKey]*analysisJob)
	}
	key := analysisKey{image: job.RegImage.String(), client: job.Registry}
	group := plan.groups[key]
	if group == nil {
		baseline, err := di.db.GetPlatformManifest(job.RegImage, job.Image.Platform)
		group = &analysisJob{baseline: baseline, err: err}
		plan.groups[key] = group
		plan.jobs = append(plan.jobs, group)
	}
	group.consumers = append(group.consumers, job)
}

func (di *Diun) runJobs(group *analysisJob) []model.NotifEntry {
	sublog := log.With().Str("image", group.consumers[0].RegImage.String()).Logger()
	entries := make([]model.NotifEntry, len(group.consumers))
	var active []int
	for i, job := range group.consumers {
		entries[i] = model.NotifEntry{
			Status:   model.ImageStatusError,
			Provider: job.Provider,
			Image:    job.RegImage,
			Metadata: job.Image.Metadata,
		}
		if !matcher.IsIncluded(job.RegImage.Tag, job.Image.IncludeTags) {
			entries[i].Status = model.ImageStatusSkip
			sublog.Debug().Str("provider", job.Provider).Msg("Tag not included")
			continue
		} else if matcher.IsExcluded(job.RegImage.Tag, job.Image.ExcludeTags) {
			entries[i].Status = model.ImageStatusSkip
			sublog.Debug().Str("provider", job.Provider).Msg("Tag excluded")
			continue
		}
		if group.err != nil {
			continue
		}
		if job.Image.AssumeImmutableTags != nil && *job.Image.AssumeImmutableTags && group.baseline.Name != "" {
			entries[i].Status = model.ImageStatusSkip
			entries[i].Manifest = group.baseline
			sublog.Debug().Str("provider", job.Provider).Msg("Skipping assumed immutable tag")
			continue
		}
		active = append(active, i)
	}
	if group.err != nil {
		sublog.Error().Err(group.err).Msg("Cannot get manifest from db")
	}
	if len(active) == 0 {
		return entries
	}

	job := group.consumers[active[0]]
	manifest, updated, err := job.Registry.Manifest(job.RegImage, group.baseline)
	if err != nil {
		if _, ok := stderrors.AsType[podmanmanifest.NonImageArtifactError](err); ok {
			for _, i := range active {
				entries[i].Status = model.ImageStatusSkip
			}
			sublog.Debug().Err(err).Msg("Skipping non-image artifact")
		} else {
			sublog.Warn().Err(err).Msg("Cannot get remote manifest")
		}
		return entries
	}

	status := model.ImageStatusUnchange
	if group.baseline.Name == "" {
		status = model.ImageStatusNew
		sublog.Info().Msg("New image found")
	} else if updated {
		status = model.ImageStatusUpdate
		sublog.Info().Msg("Image update found")
	} else {
		sublog.Debug().Msg("No changes")
	}
	for _, i := range active {
		job := group.consumers[i]
		entry := &entries[i]
		entry.Manifest = manifest
		entry.Status = status
		if status == model.ImageStatusUpdate || (status == model.ImageStatusNew && !job.FirstCheck) {
			entry.MarkUpdateAvailable()
		}
		if v, ok := manifest.Labels["org.opencontainers.image.url"]; ok {
			entry.Image.HubLink = v
		}
		if job.HubLinkOverride != "" {
			entry.Image.HubLink = job.HubLinkOverride
		}
	}
	if err := di.db.PutPlatformManifest(job.RegImage, job.Image.Platform, manifest); err != nil {
		sublog.Error().Err(err).Msg("Cannot write manifest to db")
		return entries
	}
	for _, i := range active {
		job := group.consumers[i]
		if job.FirstCheck && !*di.cfg.Watch.FirstCheckNotif {
			sublog.Debug().Str("provider", job.Provider).Msg("Skipping notification (first check)")
			continue
		}
		if !slices.Contains(job.Image.NotifyOn, model.NotifyOn(status)) {
			sublog.Debug().Str("provider", job.Provider).Msgf("Skipping notification (%s not part of specified notify status)", status)
			continue
		}
		di.notif.Send(entries[i])
	}
	return entries
}
