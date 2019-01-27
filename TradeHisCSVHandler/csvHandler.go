package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

var TIME_FROM = "2019-01-20 00:00:00 +0900 JST"
var PARSE_TYPE = "EUR/GBP"
var MAX_UNITS = "20000"
var PIPS_MULTIPLIER = 10000.0

func main(){

	// read csv
	originalCSVContent := ReadCSVFile("input/20190127.csv")

	// extract data from avaliable time
	trimedCSVContent := TrimOutUnavaliableTime(originalCSVContent)
	//Util_PrintStringArray(trimedCSVContent)

	generatedMap := GenerateOutputMap(trimedCSVContent)

	outputArray := CovertMapToArray(generatedMap)

	WriteOutputCSV(outputArray)
}

func ReadCSVFile(filePath string)([][]string)  {
	// Load a csv file.
	file, err := os.Open(filePath)

	if err != nil {
		fmt.Printf("  finished with error! \n")
		return nil
	}

	// Create a new reader.
	r := csv.NewReader(bufio.NewReader(file))

	var o [][]string

	for {
		record, err := r.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}

		//fmt.Println(record)
		//fmt.Println(len(record))

		if record[3] != PARSE_TYPE {
			continue
		}

		if record[2] == "Buy Market" ||
			record[2] == "Sell Market" ||
			record[2] == "Close Trade" ||
			record[2] == "Take Profit" ||
			record[2] == "Margin Closeout" ||
			record[2] == "Stop Loss" {
			timeUTC := record[5]
			record[5] = Util_ShowTimeJST(Util_ParseDateUTC(timeUTC))
			o = append(o, record)
		}
	}

	sort.Slice(o, func(i, j int) bool {
		return o[i][0] < o[j][0]
	})
	return o
}

func TrimOutUnavaliableTime(csvContent [][]string)([][]string)  {
	var r [][]string
	for _,s := range csvContent{
		if s[5] >= TIME_FROM {
			r = append(r, s)//fmt.Print("delete index%d \n",i)
		}
	}
	return r
}

func GenerateOutputMap(csvContent [][]string)(map[string][]string)  {
	// transactionId_in :
	// pairs, openTime, openPrice,
	// transactionId_close, closeTime, closePrice,
	// units, maxUnits, pips, pipsCovertedForTotal, profit(yen)
	o := make(map[string][]string)
	//o := map[string][]string{
	//	"transactionId_in": {"pairs",
	//		"type",
	//		"openTime",
	//		"openPrice",
	//		"transactionId_close",
	//		"closeTime",
	//		"closePrice",
	//		"units",
	//		"maxUnits",
	//		"pips",
	//		"pipsCovertedForTotal",
	//		"profit(yen)"},
	//	//"second": []string{"one", "two", "three", "four", "five"},
	//}

	for _,s := range csvContent {

		if s[2] == "Buy Market" {
			o[s[0]] = append(o[s[0]], PARSE_TYPE)	//pairs
			o[s[0]] = append(o[s[0]], "buy")	//type
			o[s[0]] = append(o[s[0]], s[5])		//openTime
			o[s[0]] = append(o[s[0]], s[6])		//openPrice
		} else if s[2] == "Sell Market" {
			o[s[0]] = append(o[s[0]], PARSE_TYPE)	//pairs
			o[s[0]] = append(o[s[0]], "sell")	//type
			o[s[0]] = append(o[s[0]], s[5])		//openTime
			o[s[0]] = append(o[s[0]], s[6])		//openPrice
		} else if s[2] == "Close Trade" ||
			s[2] == "Take Profit" ||
			s[2] == "Margin Closeout" ||
			s[2] == "Stop Loss" {
			linkedTransactionId := s[15]
			if val, ok := o[linkedTransactionId]; ok {
				o[linkedTransactionId] = append(o[linkedTransactionId], s[0])	//transactionId_close
				o[linkedTransactionId] = append(o[linkedTransactionId], s[5])	//closeTime
				closePrice := s[6]
				o[linkedTransactionId] = append(o[linkedTransactionId], closePrice)	//closePrice
				units := s[4]
				o[linkedTransactionId] = append(o[linkedTransactionId], units)	//units
				o[linkedTransactionId] = append(o[linkedTransactionId], MAX_UNITS)	//maxUnits
				openPrice := val[3]
				pips := Util_CalculatePips(openPrice, closePrice, val[1]) //pips
				o[linkedTransactionId] = append(o[linkedTransactionId], pips)
				o[linkedTransactionId] = append(o[linkedTransactionId], Util_CalculatePipsCovertedForTotal(pips, units))
				o[linkedTransactionId] = append(o[linkedTransactionId], s[9])
			}
		}
	}
	return o
}

func CovertMapToArray(outputMap map[string][]string)([][]string)  {
	var r [][]string
	for k,v := range outputMap {
		t := []string{
			k,
		}

		r = append(r, append(t, v...))
	}

	sort.Slice(r, func(i, j int) bool {
		return r[i][0] < r[j][0]
	})
	return r
}

// write CSV file
func WriteOutputCSV(outputArray [][]string)  {
	file, err := os.Create("output/result.csv")
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.WriteAll(outputArray)
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

// Util
func Util_CalculatePips(orderPrice, closePrice, tradeType string)(string)  {

	order, _ := strconv.ParseFloat(orderPrice, 64)
	close, _ := strconv.ParseFloat(closePrice, 64)

	if tradeType == "buy" {
		return fmt.Sprintf("%f", (close-order)*PIPS_MULTIPLIER)
	} else if tradeType == "sell" {
		return fmt.Sprintf("%f", (order-close)*PIPS_MULTIPLIER)
	} else {
		return ""
	}
}

func Util_CalculatePipsCovertedForTotal(pips, units string)(string)  {
	pips_f, _ := strconv.ParseFloat(pips, 64)
	units_f, _ := strconv.ParseFloat(units, 64)
	maxUnits_f, _ := strconv.ParseFloat(MAX_UNITS, 64)
	return fmt.Sprintf("%f", pips_f*units_f/maxUnits_f)
}

func Util_PrintStringArray(a [][]string)  {
	for i,s := range a {
		fmt.Printf("index: %d, content: %s\n", i, s)
	}
}

//2019-01-25 08:59:26
func Util_ParseDateUTC(dateStr string)(time.Time)  {
	layout := "2006-01-02 15:04:05"
	//str := "2019-01-25 08:59:26"
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		fmt.Print(err)
	}
	return t
}

func Util_ShowTimeJST(t time.Time)(string)  {
	tk, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		fmt.Println("err: ", err.Error())
	}
	//fmt.Print("Location:", tk, ":Time:", t.In(tk),"\n")
	return fmt.Sprintln(t.In(tk))
}


