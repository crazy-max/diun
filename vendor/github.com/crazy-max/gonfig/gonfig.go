// Lightweight config handling for Go.
//
// This library is heavily inspired by the traefik config package. It offers a simple and
// performant config handling with:
// - Configuration files in YAML or TOML format
// - Environment variables
// - Flag arguments
//
// Get started:
//
//	func init() {
//		// Typically done outside of the app
//		_ = os.Setenv("MYAPP_SERVER_FTP_HOST", "test.rebex.net")
//		_ = os.Setenv("MYAPP_SERVER_FTP_USERNAME", "demo")
//		_ = os.Setenv("MYAPP_SERVER_FTP_PASSWORD", "password")
//		_ = os.Setenv("MYAPP_SERVER_FTP_SOURCES", "/src1,/src2")
//	}
//
//	func main() {
//		cfg := Config{
//			Db: (&Db{}).GetDefaults(),
//		}
//
//		// Load from file(s)
//		fileLoader := gonfig.NewFileLoader(gonfig.FileLoaderConfig{
//			Filename: "/path/to/myapp.yml",
//			Finder: gonfig.Finder{
//				BasePaths:  []string{"/etc/myapp/myapp", "$XDG_CONFIG_HOME/myapp", "$HOME/.config/myapp", "./myapp"},
//				Extensions: []string{"yaml", "yml"},
//			},
//		})
//		if found, err := fileLoader.Load(&cfg); err != nil {
//			log.Fatal(errors.Wrap(err, "Failed to decode configuration from file"))
//		} else if !found {
//			log.Println("No configuration file found")
//		} else {
//			log.Printf("Configuration loaded from file: %s", fileLoader.GetFilename())
//		}
//
//		// Load from flags
//		flagsLoader := gonfig.NewFlagLoader(gonfig.FlagLoaderConfig{
//			Args: []string{
//				"--timezone=Europe/Paris",
//				"--logLevel=debug",
//			},
//		})
//		if found, err := flagsLoader.Load(&cfg); err != nil {
//			log.Fatal(errors.Wrap(err, "Failed to decode configuration from flags"))
//		} else if !found {
//			log.Println("No flags found")
//		} else {
//			log.Printf("Configuration loaded from flags")
//		}
//
//		// Load from environment variables
//		envLoader := gonfig.NewEnvLoader(gonfig.EnvLoaderConfig{
//			Prefix: "MYAPP_",
//		})
//		if found, err := envLoader.Load(&cfg); err != nil {
//			log.Fatal(errors.Wrap(err, "Failed to decode configuration from environment variables"))
//		} else if !found {
//			log.Println("No MYAPP_* environment variables defined")
//		} else {
//			log.Printf("Configuration loaded from %d environment variables\n", len(envLoader.GetVars()))
//		}
//
//		// Display configuration
//		b, _ := json.MarshalIndent(cfg, "", "  ")
//		fmt.Println(string(b))
//	}
package gonfig
