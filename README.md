# csv2qif
ING bank csv export file to [qif](https://en.wikipedia.org/wiki/Quicken_Interchange_Format) converter. I use this to import ING data into You Need A Budget (YNAB)

Download the zip files for a precompiled binary.

The code doesn't use any platform-specific code so you should be able to compile it on any go-supported platform.

How-To
======
- Download the csv2qif-win32zip or csv2qif-win64.zip file and extract the executable from it.
- Download your transactions as csv via the "Af- en bijschrijvingen downloaden" menu item on the left bottom of https://bankieren.mijn.ing.nl/particulier/betalen/index and save it somewhere.
- Drop the downloaded csv file on the csv2qif.exe, a new qif file with the same name should appear next to the csv file.
- Import the qif file into YNAB.
- Done!

Help
====
<pre>
This utility is for converting ING bank transaction csv files to a qif file.
The resulting qif file can be imported into You Need A Budget (YNAB)
Simple usage: Drag and drop your csv file on C:\path\to\csv2qif.exe (or run this utility with only a csv filename as argument)

Example: csv2qif.exe NL09INGB1234567890_03-10-2016_03-11-2016.csv

Advanced usage: open a commandline and use the following parameters to customize the qif output

  -i string
    	The name of the CSV file to read.
	This argument is mandatory.
  -outFile string
    	The name of the QIF file to write.
	This argument is optional, omit it to use the name of the csv file.
  -skipHeaders
    	Skip the first line of the csv file.
	This argument is optional. (default true)
  -useCode
    	Use the ING code in the qif memo.
	This argument is optional.
  -useComment
    	Use the ING comment in the qif memo.
	This argument is optional.
  -useKind
    	Use the ING transaction kind in the qif memo.
	This argument is optional. (default true)

Example: csv2qif.exe -i NL09INGB1234567890_03-10-2016_03-11-2016.csv -outFile export.qif -useCode true -useComment true
</pre>

