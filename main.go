package main

import (
	"github.com/ib407ov/CryptoCompare"
	cryptotop "github.com/ib407ov/coinmarketcup"
	"log"
	ServiceBus "main.go/servicebus"
)

func main() {

	var CryptoTop cryptotop.CoinMarketCapResponse
	CryptoTop, err := cryptotop.GetDataCryptoTop()
	if err != nil {
		error.Error(err)
	}

	data, err := ChangeStructInSlice(CryptoTop)
	if err != nil {
		error.Error(err)
	}

	var CryptoRateData cryptocompare.CryptoCompareResponseStruct

	CryptoRateData, err = GetStructCryptoRate(data)
	if err != nil {
		error.Error(err)
	}

	TopCoinsData := ChangeStructsInOneMain(CryptoTop, CryptoRateData)

	rabbit, err := ServiceBus.NewRabbitMQClient(
		"amqp://guest:guest@localhost:5672/",
		"CryptoTop",
		"crypto_top")

	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Println("RabbitMq Connect: OK")
	err = rabbit.Send(&TopCoinsData)
	log.Println("RabbitMq Send Message: OK")

	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}

}

type TopAndRate struct {
	Top    float64
	Symbol string
	Rate   float64
}

type DataTopAndRate struct {
	Data []TopAndRate
}

func (t *DataTopAndRate) GetRoutingKey() string {
	return "crypto.top"
}

func ChangeStructsInOneMain(cryptoTop cryptotop.CoinMarketCapResponse, cryptoCompare cryptocompare.CryptoCompareResponseStruct) DataTopAndRate {
	var TopCoinsData DataTopAndRate
	for _, vMarket := range cryptoTop.Data {
		for _, vCompare := range cryptoCompare.Data {
			if vMarket.Symbol == vCompare.Symbol {
				TopCoinsData.Data = append(TopCoinsData.Data, TopAndRate{
					Top:    vMarket.CmcRank,
					Symbol: vMarket.Symbol,
					Rate:   vCompare.Price,
				})
			}
		}
	}
	return TopCoinsData
}

func ChangeStructInSlice(response cryptotop.CoinMarketCapResponse) ([]string, error) {
	var cryptos []string
	for _, val := range response.Data {
		cryptos = append(cryptos, val.Symbol)
	}

	log.Println("ChangeStructInSlice: OK")
	return cryptos, nil
}

func GetStructCryptoRate(response []string) (cryptocompare.CryptoCompareResponseStruct, error) {
	data, err := cryptocompare.GetDataCurrencyRate(response)
	if err != nil {
		return cryptocompare.CryptoCompareResponseStruct{}, err
	}
	log.Println("GetStructCryptoRate: OK")

	return data, nil
}
