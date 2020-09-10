/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/IBM-Blockchain/microfab/internal/app/microfabd"
)

var logger = log.New(os.Stdout, fmt.Sprintf("[%16s] ", "microfabd"), log.LstdFlags)

func main() {
	microfabd, err := microfabd.New()
	if err != nil {
		logger.Fatalf("Failed to create application: %v", err)
	}
	err = microfabd.Start()
	if err != nil {
		logger.Fatalf("Failed to start application: %v", err)
	}
	microfabd.Wait()
}
