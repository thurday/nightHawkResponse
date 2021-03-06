/* nighthawk.audit.parser.disk.go
 * author: 0xredskull
 *
 * Disk information parser
 */


package parser 

import (
 	"fmt"
 	"os"
 	"encoding/xml"
 
 	"nighthawk/elastic"
 	nhconfig "nighthawk/config"
 	nhs "nighthawk/nhstruct"
 	nhlog "nighthawk/log"
 	nhc "nighthawk/common"
 	//nhutil "nighthawk/util"
)


func ParseDisk(caseinfo nhs.CaseInformation, auditinfo nhs.AuditType, auditfile string) {

	MAX_RECORD := nhconfig.BulkPostSize()

	xmlFile,err := os.Open(auditfile)
	if err != nil {
		nhlog.LogMessage("ParseDisk", "ERROR", fmt.Sprintf("Failed to read audit file. %s", err.Error()))
		os.Exit(nhc.ERROR_READING_AUDIT_FILE)
	}
	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	count := 0
	total := 0

	var inElement string 
	var esrecords []nhs.RlRecord

	for {
		if count == MAX_RECORD {
			elastic.ProcessOutput(caseinfo, auditinfo, esrecords)
			esrecords = esrecords[:0]
			count = 0
		}

		t,_ := decoder.Token()
		if t == nil {
			elastic.ProcessOutput(caseinfo, auditinfo, esrecords)
			esrecords = esrecords[:0]
			count = 0

			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			inElement = se.Name.Local

			if inElement == "DiskItem" {
				var item nhs.DiskItem 
				decoder.DecodeElement(&item, &se)

				item.IsWhitelisted = false

				var rlrec nhs.RlRecord
				rlrec.ComputerName = caseinfo.ComputerName
				rlrec.CaseInfo = caseinfo 
				rlrec.AuditType = auditinfo
				rlrec.Record = item 

				esrecords = append(esrecords, rlrec)
				count++
				total++
			}
		}
	}

	cmsg := fmt.Sprintf("Total DiskItem %d", total)
	nhlog.LogMessage("ParseDisk", "INFO", cmsg)
}