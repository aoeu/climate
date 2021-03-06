all :
	go run data/transformers/xmltojson.go -in data/sources/data.worldbank.org/indicator/EN.ATM.CO2E.KT/countries/en.atm.co2e.kt_Indicator_en_xml_v2.zip/en.atm.co2e.kt_Indicator_en_xml_v2.xml -out data/input/co2ekt_by_year_by_country.json -blacklist data/config/not_countries.txt -n 32

debug :
	go run data/transformers/xmltojson.go -in data/sources/data.worldbank.org/indicator/EN.ATM.CO2E.KT/countries/en.atm.co2e.kt_Indicator_en_xml_v2.zip/en.atm.co2e.kt_Indicator_en_xml_v2.xml -out data/input/co2ekt_by_year_by_country.json -debug /tmp/wat.xml 

keys :
	grep "Country or Area" data/sources/data.worldbank.org/indicator/EN.ATM.CO2E.KT/countries/en.atm.co2e.kt_Indicator_en_xml_v2.zip/en.atm.co2e.kt_Indicator_en_xml_v2.xml | sort | uniq | sed -e 's/^.*key="\(.*\)">\(.*\)<.*/\1 \2/'

serve :
	python -m SimpleHTTPServer 8086
