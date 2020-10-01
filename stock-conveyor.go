package main

import (
	"github.com/google/uuid"
	coordinatorLib "github.com/johnmillner/robo-macd/internal/coordinator"
	"github.com/johnmillner/robo-macd/internal/gatherers"
	"github.com/johnmillner/robo-macd/internal/utils"
	"log"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/common"
)

func main() {
	alpacaClient := alpaca.NewClient(common.Credentials())

	configOut := make(chan utils.Config, 100)

	coordinatorId := uuid.New()
	coordinator, mainConfigurator := coordinatorLib.InitCoordinator(configOut)

	gatherer := gatherers.InitGatherer(coordinator.NewConfigurator(gatherers.GathererConfig{
		To:         coordinatorId,
		From:       mainConfigurator.Me,
		Active:     true,
		EquityData: make(chan []gatherers.Equity, 100000),
		Client:     *alpacaClient,
		Symbols:    []string{"ABEO", "ACRX", "ACST", "GOOGL", "MAXN", "AMZN", "BAC", "TSLA", "A", "AA", "AAAU", "AACQ", "AACQU", "AADR", "AAMC", "AAN", "AAON", "AAPL", "AAT", "AAWW", "AAXN", "AB", "ABBV", "ABCB", "ABG", "ABIO", "ABMD", "ABR-A", "ABR-C", "ABTX", "ABUS", "AC", "ACA", "ACAD", "ACAM", "ACAMU", "ACB", "ACBI", "ACC", "ACCD", "ACCO", "ACEL", "ACES", "ACET", "ACEV", "ACEVU", "ACGL", "ACGLO", "ACGLP", "ACH", "ACHC", "ACHV", "ACI", "ACIA", "ACIO", "ACIU", "ACIW", "ACLS", "ACM", "ACMR", "ACN", "ACNB", "ACND", "ACND.U", "ACOR", "ACP", "ACRE", "ACRS", "ACSG", "ACSI", "ACT", "ACTCU", "ACTG", "ACU", "ACV", "ACWF", "ACWI", "ACWV", "ACWX", "ACY", "ADAP", "ADBE", "ADC", "ADCT", "ADES", "ADFI", "ADI", "ADIL", "ADM", "ADMA", "ADME", "ADMP", "ADMS", "ADNT", "ADOM", "ADP", "ADPT", "ADRE", "ADRO", "ADS", "ADSK", "ADSW", "ADT", "ADTN", "ADTX", "ADUS", "ADVM", "ADX", "ADXN", "ADXS", "ADYX", "ADZCF", "AE", "AEB", "AEE", "AEF", "AEFC", "AEG", "AEGN", "AEHR", "AEIS", "AEL", "AEL-A", "AEL-B", "AEM", "AEMD", "AEO", "AEP", "AEP-B", "AEP-C", "AER", "AERI", "AES", "AESE", "AESR", "AEY", "AEYE", "AEZS", "AFB", "AFC", "AFG", "AFGB", "AFGC", "AFGD", "AFGE", "AFGH", "AFHIF", "AFI", "AFIB", "AFIF", "AFIN", "AFINP", "AFK", "AFL", "AFLG", "AFMC", "AFMD", "AFSM", "AFT", "AFTY", "AFYA", "AG", "AGBA", "AGBAR", "AGBAU", "AGCO", "AGD", "AGE", "AGEN", "AGFS", "AGFXF", "AGG", "AGGP", "AGGY", "AGI", "AGIO", "AGLE", "AGM", "AGM-C", "AGM-D", "AGM-E", "AGM-F", "AGM.A", "AGMH", "AGNC", "AGNCN", "AGNCO", "AGNCP", "AGO", "AGO-E", "AGO-F", "AGQ", "AGR", "AGRO", "AGRX", "AGS", "AGT", "AGX", "AGYS", "AGZ", "AGZD", "AHACU", "AHC", "AHCO", "AHH", "AHH-A", "AHL-C", "AHL-D", "AHL-E", "AHPI", "AHT", "AHT-D", "AHT-G", "AHT-H", "AHT-I", "AI", "AI-B", "AI-C", "AIA", "AIC", "AIEQ", "AIF", "AIG", "AIG-A", "AIH", "AIHS", "AIIQ", "AIKI", "AIM", "AIMC", "AIMT", "AIN", "AINC", "AINV", "AIO", "AIQ", "AIR", "AIRG", "AIRI", "AIRR", "AIRT", "AIRTP", "AIT", "AIV", "AIW", "AIZ", "AIZP", "AJG", "AJRD", "AJX", "AJXA", "AKAM", "AKAOQ", "AKBA", "AKCA", "AKER", "AKO.A", "AKO.B", "AKR", "AKRO", "AKRXQ", "AKTS", "AKTX", "AKU", "AKUS", "AL", "ALAC", "ALACR", "ALACU", "ALB", "ALBO", "ALC", "ALCO", "ALDX", "ALE", "ALEC", "ALEX", "ALFA", "ALG", "ALGN", "ALGT", "ALIM", "ALIN-A", "ALIN-B", "ALIN-E", "ALJJ", "ALK", "ALKS", "ALL", "ALL-B", "ALL-G", "ALL-H", "ALL-I", "ALLE", "ALLK", "ALLT", "ALLY-A", "ALNA", "ALNY", "ALOT", "ALP-Q", "ALPN", "ALRM", "ALRN", "ALRS", "ALSK", "ALSN", "ALT", "ALTA", "ALTG", "ALTL", "ALTM", "ALTR", "ALTS", "ALTY", "ALUS", "ALUS.U", "ALV", "ALVR", "ALX", "ALXN", "ALXO", "ALYA", "AM", "AMAG", "AMAL", "AMAT", "AMBA", "AMBC", "AMBO", "AMC", "AMCA", "AMCI", "AMCIU", "AMCR", "AMCX", "AMD", "AME", "AMED", "AMEH", "AMG", "AMGN", "AMH", "AMH-D", "AMH-E", "AMH-F", "AMH-G", "AMH-H", "AMHC", "AMHCU", "AMJ", "AMK", "AMKR", "AMLP", "AMN", "AMNA", "AMNB", "AMND", "AMOM", "AMOT", "AMOV", "AMP", "AMPE", "AMPH", "AMPY", "AMRB", "AMRC", "AMRH", "AMRK", "AMRN", "AMRS", "AMRX", "AMS", "AMSC", "AMSF", "AMST", "AMSWA", "AMT", "AMTB", "AMTBB", "AMTD", "AMTI", "AMTX", "AMU", "AMUB", "AMWD", "AMWL", "AMX", "AMYT", "AMZA", "AN", "ANAB", "ANAT", "ANCN", "ANDA", "ANDAR", "ANDAU", "ANDE", "ANET", "ANF"},
		Limit:      500,
		Period:     time.Minute,
	}))

	time.Sleep(1 * time.Second)

	mainConfigurator.SendConfig(gatherers.GathererConfig{
		To:     gatherer.Configurator.Me,
		From:   mainConfigurator.Me,
		Active: false,
	})

	for simpleData := range gatherer.Configurator.Get().(gatherers.GathererConfig).EquityData {
		for _, equity := range simpleData {
			log.Printf("%v", equity)
		}
		log.Printf("%d", len(simpleData))
	}

}
