package main

import (
	"eaciit/gdrj/model"
	"eaciit/gdrj/modules"
	// "errors"
	"flag"
	"github.com/eaciit/dbox"
	_ "github.com/eaciit/dbox/dbc/mongo"
	"github.com/eaciit/toolkit"
	"os"
	// "strings"
	"sync"
	"time"
)

var (
	starttime              = time.Now()
	startperiode time.Time = time.Date(2014, 4, 1, 0, 0, 0, 0, time.UTC)
	endperiode   time.Time = startperiode.AddDate(1, 0, -1)
	// endperiode   time.Time = startperiode.AddDate(1, 0, -1)
	arroutletid []string
	arrskuid    []string
	arrbranchid []string

	//Add log variable

	arsalesmonthly []*gdrj.SalesMonthly
	mutex          = &sync.Mutex{}

	countsales, icount int
)

var conn dbox.IConnection

func setinitialconnection() {
	var err error
	conn, err = modules.GetDboxIConnection("db_godrej")

	if err != nil {
		toolkit.Println("Initial connection found : ", err)
		os.Exit(1)
	}

	err = gdrj.SetDb(conn)
	if err != nil {
		toolkit.Println("Initial connection found : ", err)
		os.Exit(1)
	}
}

func setinitialvariable() {
	fstartperiode := flag.String("startperiode", "", "start periode for selection data in format yyyy/mm/dd or dd/mm/yyyy. separated dot(.), dash(-) are permited")
	fendperiode := flag.String("endperiode", "", "end periode for selection data in format yyyy/mm/dd or dd/mm/yyyy. separated dot(.), dash(-) are permited")
	foutletid := flag.String("outletid", "", "outletid for filtering data in json string from array string")
	fskuid := flag.String("skuid", "", "skuid for filtering data in json string from array string")
	flogname := flag.String("log", "", "filename for save log")

	flag.Parse()
	_ = flogname
	if rt := modules.ToDate(toolkit.ToString(*fstartperiode)); !rt.IsZero() {
		startperiode = rt
	}

	if rt := modules.ToDate(toolkit.ToString(*fendperiode)); !rt.IsZero() {
		endperiode = rt
	}

	rstr := make([]string, 0, 0)
	if str := toolkit.ToString(*foutletid); str != "" {
		_ = toolkit.UnjsonFromString(toolkit.ToString(*foutletid), &rstr)
	}

	if len(rstr) > 0 {
		arroutletid = append(arroutletid, rstr...)
		rstr = make([]string, 0, 0)
	}

	if str := toolkit.ToString(*fskuid); str != "" {
		_ = toolkit.UnjsonFromString(toolkit.ToString(*foutletid), &rstr)
	}

	if len(rstr) > 0 {
		arrskuid = append(arrskuid, rstr...)
	}

	//set log

}

func clearpastdata() {
	//next
	return
}

func getheaderfilter() (dboxf *dbox.Filter) {
	ardboxf := make([]*dbox.Filter, 0, 0)

	ardboxf = append(ardboxf, dbox.Gte("date", startperiode))
	ardboxf = append(ardboxf, dbox.Lte("date", endperiode))
	// }

	// iarrskuid := []interface{}{}
	iarroutletid := []interface{}{}
	iarrbranchid := []interface{}{}

	for _, v := range arrbranchid {
		iarrbranchid = append(iarrbranchid, v)
	}

	for _, v := range arroutletid {
		iarroutletid = append(iarroutletid, v)
	}

	if len(arrbranchid) > 0 {
		ardboxf = append(ardboxf, dbox.In("branchid", iarrbranchid...))
	}

	if len(arroutletid) > 0 {
		ardboxf = append(ardboxf, dbox.In("outletid", iarroutletid...))
	}

	dboxf = dbox.And(ardboxf...)

	return
}

func main() {
	var mwg sync.WaitGroup

	setinitialconnection()
	setinitialvariable()
	clearpastdata()

	defer gdrj.CloseDb()
	toolkit.Println("START DATE : [", startperiode, "] END DATE : [", endperiode, "]")

	iseof := false
	dbfheader := getheaderfilter()
	cr, err := gdrj.Find(new(gdrj.SalesHeader), dbfheader, nil)
	if err != nil {
		toolkit.Println("Error Found : ", err.Error())
		os.Exit(1)
	}

	countsales = cr.Count()
	icount = 0

	defer cr.Close()
	for !iseof && countsales>0 {
		arrsalesheader := []*gdrj.SalesHeader{}
		err = cr.Fetch(&arrsalesheader, 1000, false)
		if err != nil {
			toolkit.Println("Error Found : ", err.Error())
			os.Exit(1)
		}

		if len(arrsalesheader) < 1000 {
			iseof = true
		}

		if len(arrsalesheader) == 0 {
			continue
		}

		mwg.Add(1)
		go func(tash []*gdrj.SalesHeader) {
			for _, v := range tash {
				icount += 1
				toolkit.Printfn("%d of %d data header processing in %.2f Minutes", 
					icount, countsales, time.Since(starttime).Minutes())

				vcustomer := gdrj.CustomerGetByID(toolkit.Sprintf("%v%v", v.BranchID, v.OutletID))

				dbfdetail := dbox.Eq("salesheaderid", v.ID)
				gcr, err := gdrj.Find(new(gdrj.SalesDetail), dbfdetail, nil)
				if err != nil {
					toolkit.Println("Error Found : ", err.Error())
					os.Exit(1)
				}

				arrsalesdetail := []*gdrj.SalesDetail{}
				err = gcr.Fetch(&arrsalesdetail, 0, false)
				if err != nil {
					toolkit.Println("Error Found : ", err.Error())
					os.Exit(1)
				}

				allocdiscount := v.SalesDiscountAmount / toolkit.ToFloat64(cr.Count(), 0, toolkit.RoundingAuto)

				for _, xv := range arrsalesdetail {
					xvproduct := gdrj.ProductGetBySKUID(xv.SKUID_SAPBI)
					xvprofitcenter := gdrj.ProfitCenterGetByID(toolkit.Sprintf("%v%v", v.BranchID, xvproduct.BrandCategoryID))
					// 70000000	SALES -  TRADE DOMESTIC // 70000302	SALES RETURN -V-DIST // 75053730	SALES DISCOUNT REGULAR
					mutex.Lock()
					xvls := new(gdrj.LedgerSummary)
					if xv.SalesGrossAmount < 0 {
						xvls = gdrj.GetLedgerSummaryByDetail("70000302", xvprofitcenter.ID, "", vcustomer.ID, xvproduct.ID, v.Date.Year(), v.Date.Month())
						xvls.LedgerAccount = "70000302"
						xvls.PLCode = "PL01"
						xvls.PLOrder = "PL0001"
					} else {
						xvls = gdrj.GetLedgerSummaryByDetail("70000000", xvprofitcenter.ID, "", vcustomer.ID, xvproduct.ID, v.Date.Year(), v.Date.Month())
						xvls.LedgerAccount = "70000000"
						xvls.PLCode = "PL04"
						xvls.PLOrder = "PL0004"
					}
					xvls.Value1 += xv.SalesGrossAmount

					if xvls.ID == "" {
						xvls.PC = xvprofitcenter
						xvls.CompanyCode = "ID11"
						xvls.Customer = vcustomer
						xvls.Product = xvproduct
						xvls.Date = &gdrj.Date{
							Month: v.Date.Month(),
							Year:  v.Date.Year()}
						xvls.PCID = xvprofitcenter.ID
						xvls.OutletID = vcustomer.ID
						xvls.SKUID = xvproduct.ID
						xvls.Month = v.Date.Month()
						xvls.Year = v.Date.Year()
					}

					xvls.ID=xvledgersummarydisc.PrepareID().(string)
					err = gdrj.Save(xvls)
					if err != nil {
						toolkit.Println("Error Found : ", err.Error())
						os.Exit(1)
					}

					xvledgersummarydisc := gdrj.GetLedgerSummaryByDetail("75053730", xvprofitcenter.ID, "", vcustomer.ID, xvproduct.ID, v.Date.Year(), v.Date.Month())
					xvledgersummarydisc.LedgerAccount = "75053730"
					xvledgersummarydisc.PLCode = "PL07"
					xvledgersummarydisc.PLOrder = "PL0007"
					xvledgersummarydisc.Value1 += allocdiscount
					if xv.SalesGrossAmount > xv.SalesNetAmount {
						xvledgersummarydisc.Value1 += (xv.SalesGrossAmount - xv.SalesNetAmount)
					}

					if xvledgersummarydisc.ID == "" {
						xvledgersummarydisc.PC = xvprofitcenter
						xvledgersummarydisc.CompanyCode = "ID11"
						xvledgersummarydisc.Customer = vcustomer
						xvledgersummarydisc.Product = xvproduct
						xvledgersummarydisc.Date = &gdrj.Date{
							Month: v.Date.Month(),
							Year:  v.Date.Year()}
						xvledgersummarydisc.PCID = xvprofitcenter.ID
						xvledgersummarydisc.OutletID = vcustomer.ID
						xvledgersummarydisc.SKUID = xvproduct.ID
						xvledgersummarydisc.Month = v.Date.Month()
						xvledgersummarydisc.Year = v.Date.Year()
					}

					if allocdiscount > 0 || (xv.SalesGrossAmount-xv.SalesNetAmount) > 0 {
						xvledgersummarydisc.ID=xvledgersummarydisc.PrepareID().(string)
						err = gdrj.Save(xvledgersummarydisc)
						if err != nil {
							toolkit.Println("Error Found : ", err.Error())
							os.Exit(1)
						}
					}
					mutex.Unlock()
				}

				gcr.Close()
			}
			mwg.Done()
		}(arrsalesheader)
	}

	mwg.Wait()
}
