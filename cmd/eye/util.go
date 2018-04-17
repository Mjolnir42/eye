/*
Copyright (c) 2016, Jörg Pernfuß <code.jpe@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/mjolnir42/soma/lib/proto"
	"github.com/satori/go.uuid"
)

func abortOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 200
func dispatchJSONOK(w *http.ResponseWriter, jsonb *[]byte) {
	(*w).Header().Set("Content-Type", "application/json")
	(*w).WriteHeader(http.StatusOK)
	(*w).Write(*jsonb)
}

// 204
func dispatchNoContent(w *http.ResponseWriter) {
	(*w).WriteHeader(http.StatusNoContent)
	(*w).Write(nil)
}

// 400
func dispatchBadRequest(w *http.ResponseWriter, err string) {
	http.Error(*w, err, http.StatusBadRequest)
	log.Println(err)
}

// 404
func dispatchNotFound(w *http.ResponseWriter) {
	http.Error(*w, "No items found", http.StatusNotFound)
	log.Println("No items found")
}

// 410
func dispatchGone(w *http.ResponseWriter, err string) {
	http.Error(*w, err, http.StatusGone)
	log.Println(err)
}

// 412
func dispatchPrecondition(w *http.ResponseWriter, err string) {
	http.Error(*w, err, http.StatusPreconditionFailed)
	log.Println(err)
}

// 422
func dispatchUnprocessable(w *http.ResponseWriter, err string) {
	http.Error(*w, err, 422)
	log.Println(err)
}

// 500
func dispatchInternalServerError(w *http.ResponseWriter, err string) {
	http.Error(*w, err, http.StatusInternalServerError)
	if Eye.Volatile {
		log.Fatal(err)
	}
	log.Println(err)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
