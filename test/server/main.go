////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"net/http"
)

func main() {
	port := "9090"
	root := "../assets"
	fmt.Printf("Starting server on port %s from %s\n", port, root)

	err := http.ListenAndServe(":"+port, http.FileServer(http.Dir(root)))
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
