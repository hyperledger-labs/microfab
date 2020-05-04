/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"log"

	"github.com/IBM-Blockchain/fablet/internal/app/fablet"
)

func main() {
	fablet, err := fablet.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	err = fablet.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
