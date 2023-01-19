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
	root := ""
	fmt.Printf("Starting server on port %s from %s\n", port, root)
	fmt.Printf("\thttp://localhost:%s\n", port)

	err := http.ListenAndServe(":"+port, http.FileServer(http.Dir(root)))
	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %+v", err))
	}
}
