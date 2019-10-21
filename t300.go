package main

import (
	"log"
	//"strings"
	"io/ioutil"
	"flag"
	"strconv"
	"text/template"
	"os"
	copy "github.com/otiai10/copy"
	"github.com/tealeg/xlsx"
	"fmt"
)

// Configuration data for a single RTU
type RTUConf struct {
	Name string
	StreetAddress string
	CommonAddress int
	IP string
	Netmask string
	DefaultGW string
	SNTPServer string
	ProtConfs []ProtConf
}

// Configuration data for a single protection
type ProtConf struct {
	Name string
	Num int
	IP string
	Netmask string
	DefaultGW string
	Affacciata string
}

func main() {
	// array of configurations from input file
	configurations := make(map[string]*RTUConf)
	// Template projects
	confDirs := map[int]string {
		1: "Config_1_REF_V3_TEMPLATE",
		2: "Config_2_REF_V3_TEMPLATE", 
		3: "Config_3_REF_V3_TEMPLATE",
	}
	// input configuration file
	var conf = flag.String("conf", "conf.xlsx", "Input configuration file")
	// template directory path
	var templatePath = flag.String("tmpl", ".", "Path to template projects")
	flag.Parse()
	
	// parse the configuration file
	xlFile, err := xlsx.OpenFile(*conf)
	if err != nil {
		log.Fatal(err)
	}
	// RTU sheet
	rtuSheet := xlFile.Sheet["RTU"]
	for i, row := range rtuSheet.Rows {
		newConf := RTUConf {};
		// skip header line
		if i == 0 {
			continue
		}
		for i, cell := range row.Cells {
			switch i {
			case 0:
				newConf.Name = cell.String() 
			case 1:
				newConf.StreetAddress = cell.String()
			case 2:
				ca, err := strconv.Atoi(cell.String())
				if err != nil {
					continue
				}	
				newConf.CommonAddress = ca
			case 3:
				newConf.IP = cell.String()
			case 4:
				newConf.Netmask = cell.String()
			case 5:
				newConf.DefaultGW = cell.String()
			case 6:
				newConf.SNTPServer = cell.String()
			}
		}
		configurations[newConf.Name] = &newConf
        }
	// PROTEZIONI sheet
	protSheet := xlFile.Sheet["PROTEZIONI"]
	for i, row := range protSheet.Rows {
		var rtu *RTUConf
		newProt := ProtConf {};
		// skip header line
		if i == 0 {
			continue
		}
		for i, cell := range row.Cells {
			switch i {
			case 0:
				rtu = configurations[cell.String()]
			case 1:
				num, err := strconv.Atoi(cell.String())
				if err != nil {
					continue
				}
				newProt.Num = num
			case 2:
				newProt.Name = cell.String()
			case 3:
				newProt.IP = cell.String()
			case 4:
				newProt.Netmask = cell.String()
			case 5:
				newProt.DefaultGW = cell.String()
			case 6:
				newProt.Affacciata = cell.String()
			}
		}
		rtu.ProtConfs = append(rtu.ProtConfs, newProt)
        }
	// for each configuration generate the project from template
	for _, configuration := range configurations {
		tmplDir := confDirs[len(configuration.ProtConfs)]
		err = copy.Copy(*templatePath + "/" + tmplDir, configuration.Name)
		if err != nil {
			log.Fatal(err)
		}
		// parse Profile.xml
		profilePath := configuration.Name + "/" + tmplDir + " Files" + "/Profile.xml"
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
