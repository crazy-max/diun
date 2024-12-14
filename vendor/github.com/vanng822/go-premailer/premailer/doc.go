// Package premailer is for inline styling.
//
// 	import (
// 		"fmt"
// 		"github.com/vanng822/go-premailer/premailer"
// 		"log"
//	)
//
//	func main() {
//		prem, err := premailer.NewPremailerFromFile(inputFile, premailer.NewOptions())
//		if err != nil {
//			log.Fatal(err)
//		}
//		html, err := prem.Transform()
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Println(html)
//	}
//	// Input
//
//	<html>
//	<head>
//	<title>Title</title>
//	<style type="text/css">
//		h1 { width: 300px; color:red; }
//		strong { text-decoration:none; }
//	</style>
//	</head>
//	<body>
//		<h1>Hi!</h1>
//		<p><strong>Yes!</strong></p>
//	</body>
//	</html>
//
// // Output
//
//	<html>
//	<head>
//	<title>Title</title>
//	</head>
//	<body>
//		<h1 style="color:red;width:300px" width="300">Hi!</h1>
//		<p><strong style="text-decoration:none">Yes!</strong></p>
//	</body>
//	</html>
package premailer
