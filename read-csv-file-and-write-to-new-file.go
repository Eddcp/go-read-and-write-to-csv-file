package main

import (
    "encoding/csv"
	"encoding/json"
    "log"
	"fmt"
    "os"
	"net/http"
	"io/ioutil"
	"sync"
)

var wg sync.WaitGroup
var URL = "https://swapi.dev/api/"

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func readCSVFile(filePath string) [][]string {
	csvFile, err := os.Open(filePath)
	checkError("Error while opening file", err)

	fmt.Println("Successfully Opened CSV file")
	defer csvFile.Close()

	/*
		Be careful when using ReadAll, if file is to big this is a bad way to store all data on memory
		Maybe in the future change this function to write to file as the program reads each line
	*/
	csvLines, err := csv.NewReader(csvFile).ReadAll()
    checkError("Error while reading file", err)

	return csvLines
}

func writeFileWithNewColumn(filePath string, csvLines [][]string, columnName string) error {
	length := len(csvLines)
	ch := make(chan map[string]interface{})
	
	csvLines[0] = append(csvLines[0], columnName)
	for i := 1; i < length; i++ {
		wg.Add(1)
		go callAPI(URL+csvLines[i][1]+"/"+csvLines[i][0], ch)
		data := <-ch
		csvLines[i] = append(csvLines[i], getResourceName(data))
	}

	//close the channel in the background with immediate invocation of function
	go func() {
		wg.Wait()
		close(ch)
	}()

	writeFile, err := os.Create(filePath)
	checkError("Error while opening write file", err)
	writer := csv.NewWriter(writeFile)
	if err = writer.WriteAll(csvLines); err != nil {
		writeFile.Close()
		return err
	}

	return writeFile.Close()


}

func getResourceName(data map[string]interface{}) string {
	resourceName := data["name"]
	if resourceName != nil {
		return resourceName.(string)
	}
	return "None"
}

func callAPI(url string, ch chan<- map[string]interface{}) {
	defer wg.Done()
	response, err := http.Get(url)
	checkError("Error calling API", err)
	body, err := ioutil.ReadAll(response.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	ch <- result
}

func main() {

	file := os.Args[1]
	csvLines := readCSVFile(file)

	if err := writeFileWithNewColumn("./writeFile.csv", csvLines, "resource_name"); err != nil {
		checkError("Error processing file and trying to add new column", err)
	}

}