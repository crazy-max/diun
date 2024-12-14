// Go client library for accessing the Healthchecks API.
//
// Get started:
//  func main() {
//    var err error
//    client := gohealthchecks.NewClient(nil)
//
//    err = client.Start(context.Background(), gohealthchecks.PingingOptions{
//      UUID: "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278",
//      Logs: "Job started!",
//    })
//    if err != nil {
//      log.Fatal(err)
//    }
//
//    err = client.Success(context.Background(), gohealthchecks.PingingOptions{
//      UUID: "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278",
//      Logs: "Job completed!",
//    })
//    if err != nil {
//      log.Fatal(err)
//    }
//
//    err = client.Fail(context.Background(), gohealthchecks.PingingOptions{
//      UUID: "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278",
//      Logs: "Job failed...",
//    })
//    if err != nil {
//      log.Fatal(err)
//    }
//  }
package gohealthchecks
