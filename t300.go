package main

import (
	"encoding/csv"
	"log"
	"strings"
	"io/ioutil"
	"flag"
	"strconv"
	"text/template"
	"os"
	copy "github.com/otiai10/copy"
)

// Configuration data for a single RTU
type Configuration struct {
	RTU string
	NumLines int
	IP string
}

func main() {
	// array of configurations from input file
	var configurations []Configuration
	// Template projects
	confDirs := map[int]string {
		1: "Config_1_REF_V3_TEMPLATE",
		2: "Config_2_REF_V3_TEMPLATE", 
		3: "Config_3_REF_V3_TEMPLATE",
	}
	// input configuration file
	var conf = flag.String("conf", "conf.csv", "Input configuration file")
	// template directory path
	var templatePath = flag.String("tmpl", ".", "Path to template projects")
	flag.Parse()
	
	// parse the configuration file
	csvIn, _ := ioutil.ReadFile(*conf)
	r := csv.NewReader(strings.NewReader(string(csvIn)))
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	for i, record := range records {
		// skip header line
		if i == 0 {
			continue
		}
		numLines, err := strconv.Atoi(record[1])
		if err != nil {
			continue
		}
		newConf := Configuration {
			RTU: record[0],
			NumLines: numLines,
			IP: record[2],
		}
		configurations = append(configurations, newConf)
	}
	// for each configuration generate the project from template
	for _, configuration := range configurations {
		tmplDir := confDirs[configuration.NumLines]
		err = copy.Copy(*templatePath + "/" + tmplDir, configuration.RTU)
		if err != nil {
			log.Fatal(err)
		}
		// parse Profile.xml
		profilePath := configuration.RTU + "/" + tmplDir + " Files" + "/Profile.xml"
		data, err := ioutil.ReadFile(profilePath)
		if err != nil {
			log.Fatal(err)
		}
		profileFile, err := os.OpenFile(profilePath, os.O_WRONLY, 0644) 
		if err != nil {
			log.Fatal(err)
		}
		tmpl, err := template.New("test").Parse(string(data))
		if err != nil {
			log.Fatal(err)
		}
		
		err = tmpl.Execute(profileFile, configuration)
		if err != nil {
			log.Fatal(err)
		}
		// close the file
		if err := profileFile.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
