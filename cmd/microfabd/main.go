/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"log"

	"github.com/IBM-Blockchain/microfab/internal/app/microfabd"
)

func main() {
	microfabd, err := microfabd.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	err = microfabd.Start()
	if err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
	microfabd.Wait()
}
