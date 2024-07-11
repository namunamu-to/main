package main

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
)

func MakeUuid() string {
	id := uuid.New()
	return id.String()
}

func ReadFileStr(filePath string) string {
	// ファイルを開く
	f, err := os.Open(filePath)
	if err != nil {
		println(err)
	}

	defer f.Close()

	//ファイルをbyteで読み、stringに変換。
	buf, err := io.ReadAll(f)
	if err != nil {
		println(err)
	}
	result := string(buf)

	return result
}

func ReadFileLine(filePath string) []string {
	result := strings.Split(ReadFileStr(filePath), "\n")
	return result
}

func ReadCsv(filePath string) [][]string {
	return CsvToSlice(ReadFileStr(filePath))
}

func CsvToSlice(csvStr string) [][]string {
	csvLine := strings.Split(csvStr, "\n")

	var result [][]string
	for i := 0; i < len(csvLine); i++ {
		result = append(result, strings.Split(string(csvLine[i]), ","))
	}

	return result
}

func SliceToCsvStr(csv [][]string) string {
	csvLines := []string{}
	for _, csvLine := range csv {
		csvLines = append(csvLines, strings.Join(csvLine, ","))

	}

	return strings.Join(csvLines, "\n")
}

func WriteCsv(filepath string, csv [][]string) {
	// csvLines := []string{}
	// for _, csvLine := range csv {
	// 	csvLines = append(csvLines, strings.Join(csvLine, ","))

	// }

	// WriteFile(filepath, strings.Join(csvLines, "\n"))
	WriteFile(filepath, SliceToCsvStr(csv))
}

func JsonToMap(jsonString string) (map[string]string, error) {
	var data map[string]string
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		println(err.Error())
	}

	return data, nil
}

type msgJson struct {
	Category, Msg string
}

func strToMsgJson(jsonString string) (msgJson, error) {
	var msg msgJson
	err := json.Unmarshal([]byte(jsonString), &msg)
	return msg, err
}

func WriteFile(filePath string, writeStr string) {
	// 1. 書き込み先のファイル作成
	f, err := os.Create(filePath)
	if err != nil {
		println(err)
	}

	defer f.Close()

	// 2. バイト文字列に変換
	d := []byte(writeStr)

	// 3. 書き込み
	f.Write(d)
}

func WriteFileAppend(filepath string, writeStr string) {
	fileStr := ReadFileStr(filepath)
	fileStr += "\n" + writeStr
	WriteFile(filepath, fileStr)
}

func IsFile(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
