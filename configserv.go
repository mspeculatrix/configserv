/*
Package main
Application: configserv
Runs on the Raspberry Pi on board a robot. It's used to tell the RPi
the location/configuration of the remote server used to, for example,
log telemetry or other data. The idea is that, after starting, the
user will employ a web app running on the remote server to send this
info and then the RPi will configure its settings.

Might be expandable to handle other configurations and even act as
some kind of REST API for robot.

It receives a GET request via HTTP and converts the query string to
a map before saving the map as key/value pairs in a config file,
using the format k=v.

Offered up under GPL 3.0 but absolutely not guaranteed fit for use.
This is code created by an amateur dilettante, so use at your own risk.
Github: https://github.com/mspeculatrix
Blog: https://mansfield-devine.com/speculatrix/
*/

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/mspeculatrix/msgolib/codeutils"
	"github.com/mspeculatrix/msgolib/fileutils"
	"github.com/mspeculatrix/msgolib/webutils"
)

// *** DEFAULT SETTINGS ***
var verbose = false            // modify with -v
var robotName = "robot"        // used in returned headers
var portNum = "3000"           // port num for this server
var cfgFile = "remoteSvr.cfg"  // file to store received config data
var logFile = "configserv.log" // for logging, duh

// We'll store the PID of this process/program in a file. This is
// intended for use by monitoring programs.
const pidFile = "configserv.pid" // to store PID for this program

// fatalErr is used to both log fatal errors and output to screen
func fatalErr(msg string, err error) {
	fmt.Println(msg, err)
	log.Fatalln(msg, err)
}

/******************************************************************************
 *****   HTTP FUNCTIONS                                                   *****
 ******************************************************************************/
// addStandardHeaders() - outputs HTTP headers we send with every response.
func addStandardHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Server", "configserv")
	w.Header().Set("X-Robot-Name", robotName)
}

// handleDefault() - default/fallback handler for all routes not matched
// by other handlers.
func handleDefault(w http.ResponseWriter, r *http.Request) {
	addStandardHeaders(w)
	w.WriteHeader(http.StatusNotFound) // send 404
	fmt.Fprintf(w, "unknown")
}

// handleRemoteConfig() - HTTP handler for /remcfg route
func handleRemoteConfig(w http.ResponseWriter, r *http.Request) {
	params, err := webutils.SimpleParseQuery(r.URL.String())
	if err != nil {
		return
	}
	params["recvdFrom"] = r.RemoteAddr // add IP:port of sender to config
	if _, err = fileutils.WriteConfigFile(cfgFile, params); err != nil {
		log.Println("*** Could not write to config file:", err, "***")
		return
	}
	log.Println("- wrote config file:", cfgFile)
	if verbose {
		codeutils.PrintStringMap(&params)
	}
	// send response
	addStandardHeaders(w)
	w.WriteHeader(http.StatusOK) // send 200
	fmt.Fprintf(w, "OK")
}

/******************************************************************************
 *****   MAIN                                                             *****
 ******************************************************************************/
func main() {
	// ===== READ FLAGS =======================================================
	flag.StringVar(&cfgFile, "f", cfgFile, "config filename (with full path)")
	flag.StringVar(&logFile, "l", logFile, "log file filename (with full path)")
	flag.StringVar(&robotName, "n", robotName, "robot name")
	flag.StringVar(&portNum, "p", portNum, "port to run on")
	flag.BoolVar(&verbose, "v", verbose, "produce verbose output")
	flag.Parse()

	// Check that config file is writeable. If not, exit.
	fh, err := os.Create(cfgFile)
	if err != nil {
		fatalErr("Could not create configuration file", err)
	}
	fh.Close()

	// ===== SET UP LOGGING ===================================================
	fhlog, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fatalErr("Failed to open log file "+logFile+":", err)
	}
	log.SetOutput(io.Writer(fhlog)) // use our log file for logging
	log.Println("Running configserv")
	log.Printf("- on port   : %s\n", portNum)
	log.Printf("- robot name: %s\n", robotName)

	// ===== PID =============================================================
	oldPid, err := fileutils.ReadPIDFile(pidFile)
	if err != nil {
		log.Println(err)
	} else if oldPid != "" {
		log.Printf("- previous PID found: %s\n", oldPid)
	}
	if _, err = fileutils.WritePIDToFile(os.TempDir() + "/" + pidFile); err != nil {
		log.Println("- could not write to temp PID file:", err)
		// maybe the temp dir wasn't writable. Try again using just
		// a local file
		if _, err = fileutils.WritePIDToFile(pidFile); err != nil {
			fatalErr("*** Could not write to PID file ***", err)
		}
		log.Println("- created local PID file:", pidFile)
	} else {
		log.Println("- created PID file:", os.TempDir()+"/"+pidFile)
	}

	// ===== HTTP SERVER ======================================================
	// Route handlers

	// REMOTE CONFIG
	http.HandleFunc("/remcfg", handleRemoteConfig)

	// DEFAULT
	// Following line must be the last of the handlers as it
	// catches anything not dealt with above.
	http.HandleFunc("/", handleDefault)

	// Start listening. The use of nil for the handler means
	// the DefaultServeMux (set using the HandleFunc() calls
	// above) is used.
	http.ListenAndServe(":"+portNum, nil)
}
