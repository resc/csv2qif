package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	useCode     bool
	useKind     bool
	useComment  bool
	skipHeaders bool
	inFile      string
	outFile     string
	write       func(s string) = func(s string) {
		print(s)
	}
	writeLn func(s string) = func(s string) {
		write(s)
		write("\n")
	}
)

const (
	QifHeader string = "!Type:Bank"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "This utility is for converting ING bank transaction csv files to a qif file.\n"+
			"The resulting qif file can be imported into You Need A Budget (YNAB)\n"+
			"Simple usage: Drag and drop your csv file on %s (or run this utility with only a csv filename as argument)\n"+

			"\nExample: csv2qif.exe NL09INGB1234567890_03-10-2016_03-11-2016.csv\n\n"+
			"Advanced usage: open a commandline and use the following parameters to customize the qif output\n\n", os.Args[0])

		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample: csv2qif.exe -i NL09INGB1234567890_03-10-2016_03-11-2016.csv -outFile export.qif -useCode true -useComment true\n\n")
	}

	flag.StringVar(&inFile, "i", "", "The name of the CSV file to read.\n\tThis argument is mandatory.")
	flag.StringVar(&outFile, "outFile", "", "The name of the QIF file to write.\n\tThis argument is optional, omit it to use the name of the csv file.")
	flag.BoolVar(&skipHeaders, "skipHeaders", true, "Skip the first line of the csv file.\n\tThis argument is optional.")
	flag.BoolVar(&useCode, "useCode", false, "Use the ING code in the qif memo.\n\tThis argument is optional.")
	flag.BoolVar(&useKind, "useKind", true, "Use the ING transaction kind in the qif memo.\n\tThis argument is optional.")
	flag.BoolVar(&useComment, "useComment", false, "Use the ING comment in the qif memo.\n\tThis argument is optional.")

	flag.Parse()
}

func main() {
	defer func() {
		err, ok := recover().(error)
		if ok {
			// print the error and sleep 10 seconds,
			// so the draggie-droppie types can read the error too...
			println(err)
			time.Sleep(10 * time.Second)
		}
	}()
	run()
}

func run() {

	// drag-drop support, explorer starts the exe with only the full file path
	// as the first argument if you drag-drop the file...
	if len(os.Args) == 2 && len(inFile) == 0 {
		inFile = os.Args[1]
	}

	fileInfo, err := os.Stat(inFile)

	if os.IsNotExist(err) {
		flag.Usage()
		return
	} else if err != nil {
		flag.Usage()
		log.Fatalf("Error opening  input file '%s': %s", inFile, err.Error())
	}

	if fileInfo.Size() < 1 {
		flag.Usage()
		log.Fatalf("Input file '%s' is empty", inFile)

	}

	// open input file
	input, err := os.Open(inFile)
	if err != nil {
		flag.Usage()
		log.Fatalf("Error opening  input file '%s': %s", inFile, err.Error())
	}
	defer input.Close()

	// open output file
	if len(outFile) == 0 {
		outFile = inFile[:len(inFile)-3] + "qif"
	}

	output, err := os.Create(outFile)
	if err != nil {
		flag.Usage()
		log.Fatalf("Error opening  output file '%s': %s", outFile, err.Error())
	}
	defer output.Close()

	// redefine write to write to file
	write = func(s string) {
		_, err := output.WriteString(s)
		if err != nil {
			flag.Usage()
			log.Fatalf("Error writing file '%s': %s", outFile, err.Error())
		}
	}

	line := 0
	rdr := csv.NewReader(input)

	// write the Qif header
	writeLn(QifHeader)
	for {
		line += 1

		raw, err := rdr.Read()
		if err == io.EOF {
			// end of file
			break
		}

		if err != nil {
			// csv read error
			flag.Usage()
			log.Fatalf("line %d of %s has an error: %s", line, inFile, err.Error())
		}

		if line == 1 && skipHeaders {
			// skip the csv headers
			continue
		}

		// parse record from csv.
		record, err := ParseRecord(raw)
		if err != nil {
			flag.Usage()
			log.Fatalf("line %d of %s has an error: %s", line, inFile, err.Error())
		}

		// write record to output
		write(record.QifRecord())
	}

	fmt.Printf("Wrote %s", outFile)

}

type (
	// Record holds the fields of the csv import file by name.
	Record struct {
		// Date hold the transaction date
		Date time.Time
		// Name hold the name of the account holder or account description
		Name string
		// IBAN holds the account IBAN number
		IBAN string
		// OtherIBAN holds the other account IBAN number, if applicable.
		OtherIBAN string
		// Code contains the transaction code
		Code string
		// Direction contains the transaction direction "Af" means from IBAN to OtherIban, "Bij" means from OtherIBAN to IBAN
		Direction string
		// Amount is the amount of the transaction in euros
		Amount float64
		// TransactionKind holds the transaction kind (this corresponds with Code)
		TransactionKind string
		// Comments holds the transaction comments
		Comments string
		// RawFields holds the fields as read from the csv file
		RawFields []string
	}
)

// QifRecord returns the record as a qif formatted record string
func (r *Record) QifRecord() string {

	comments := make([]string, 0, 3)
	if useCode {
		comments = append(comments, r.Code)
	}
	if useKind {
		comments = append(comments, r.TransactionKind)
	}
	if useComment {
		comments = append(comments, r.Comments)
	}

	return fmt.Sprintf("D%s\n", r.Date.Format("01/02/2006")) +
		fmt.Sprintf("T%.2f\n", r.Amount) +
		fmt.Sprintf("U%.2f\n", r.Amount) +
		fmt.Sprintf("P%s\n", r.Name) +
		fmt.Sprintf("M%s\n", strings.Join(comments, " ")) +
		fmt.Sprintf("Cc\n") + // cleared status, always cleared.
		fmt.Sprintf("N\n") +
		"^\n"
}

// ParseRecord parses a csv Record.
func ParseRecord(raw []string) (*Record, error) {
	if len(raw) != 9 {
		return nil, fmt.Errorf("Wrong number of columns in record, expected 9, got %d", len(raw))
	}
	date, err := time.Parse("20060102", raw[0])
	if err != nil {
		return nil, fmt.Errorf("Error parsing date: %s", err.Error())
	}
	amount, err := strconv.ParseFloat(strings.Replace(raw[6], ",", ".", 1), 64)
	if err != nil {
		return nil, fmt.Errorf("Error parsing amount:  %s", err.Error())

	}

	r := &Record{
		Date:            date,
		Name:            strings.TrimSpace(raw[1]),
		IBAN:            strings.TrimSpace(raw[2]),
		OtherIBAN:       strings.TrimSpace(raw[3]),
		Code:            strings.TrimSpace(raw[4]),
		Direction:       strings.TrimSpace(raw[5]),
		Amount:          amount,
		TransactionKind: strings.TrimSpace(raw[7]),
		Comments:        strings.TrimSpace(raw[8]),
		RawFields:       raw,
	}

	if strings.Contains(r.Direction, "Af") {
		r.Amount = -math.Abs(amount)
	}

	return r, nil
}
