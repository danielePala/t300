package main

import (
	"flag"
	"fmt"
	"github.com/otiai10/copy"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"text/template"
)

// Configuration data for a single RTU
type RTUConf struct {
	Name          string
	StreetAddress string
	CommonAddress int
	IP            string
	Netmask       string
	DefaultGW     string
	SNTPServer    string
	ProtConfs     []ProtConf
}

// Configuration data for a single protection
type ProtConf struct {
	Name       string
	Num        int
	IP         string
	Netmask    string
	DefaultGW  string
	Affacciata string
}

func main() {
	// Template projects
	confDirs := map[int]string{
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
	configurations := parseRTU(xlFile.Sheet["RTU"])
	// PROTEZIONI sheet
	parseProtections(xlFile.Sheet["PROTEZIONI"], configurations)

	// for each configuration generate the project from template
	for _, configuration := range configurations {
		tmplDir := confDirs[len(configuration.ProtConfs)]
		err = copy.Copy(*templatePath+"/"+tmplDir, configuration.Name)
		if err != nil {
			log.Fatal(err)
		}
		prjPath := configuration.Name + "/" + configuration.Name
		prjFilesPath := prjPath + " Files"
		// rename paths according to generated project
		err = os.Rename(configuration.Name+"/"+tmplDir+".ctpx", prjPath+".ctpx")
		err = os.Rename(configuration.Name+"/"+tmplDir+" Files", prjFilesPath)

		// parse Profile.xml
		profilePath := prjFilesPath + "/Profile.xml"
		parseTemplate(profilePath, configuration)
		// parse T300_61850.scd
		scdPath := prjFilesPath + "/i61sc/T300_61850.scd"
		parseTemplate(scdPath, configuration)
		// parse i4e_cont.xml
		i4ePath := prjFilesPath + "/i4e/i4e_cont.xml"
		parseTemplate(i4ePath, configuration)
		// parse thmConf.xml
		thmPath := prjFilesPath + "/thmConf.xml"
		parseTemplate(thmPath, configuration)
	}
}

func parseRTU(rtuSheet *xlsx.Sheet) map[string]*RTUConf {
	// array of configurations from input file
	configurations := make(map[string]*RTUConf)
	for i, row := range rtuSheet.Rows {
		newConf := RTUConf{}
		// skip header line
		if i == 0 {
			continue
		}
		if len(row.Cells[0].String()) == 0 {
			break
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
		fmt.Println("Added new RTU: ", newConf.Name)
	}
	return configurations
}

func parseProtections(protSheet *xlsx.Sheet, configurations map[string]*RTUConf) {
	for i, row := range protSheet.Rows {
		var rtu *RTUConf
		newProt := ProtConf{}
		// skip header line
		if i == 0 {
			continue
		}
		if len(row.Cells[0].String()) == 0 {
			break
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
		fmt.Printf("Added new Protection %s to RTU %s\n", newProt.Name, rtu.Name)
		rtu.ProtConfs = append(rtu.ProtConfs, newProt)
	}
}

func parseTemplate(fileName string, configuration *RTUConf) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(fileName, os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	tmpl, err := template.New("test").Parse(string(data))
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(file, configuration)
	if err != nil {
		log.Fatal(err)
	}
	// close the file
	if err := file.Close(); err != nil {
		log.Fatal(err)
	}
}
