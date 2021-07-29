package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mcktr/check_fritz/modules/fritz"
	"github.com/mcktr/check_fritz/modules/perfdata"
	"github.com/mcktr/check_fritz/modules/thresholds"
)

// CheckDownstreamMax checks the maximum downstream that is available on this internet connection
func CheckDownstreamMax(aI ArgumentInformation) {
	resps := make(chan []byte)
	errs := make(chan error)

	var soapReq fritz.SoapData

	isDSL := false

	if strings.ToLower(*aI.Modelgroup) == "dsl" {
		isDSL = true
	}

	if isDSL {
		soapReq = fritz.CreateNewSoapData(*aI.Username, *aI.Password, *aI.Hostname, *aI.Port, "/upnp/control/wandslifconfig1", "WANDSLInterfaceConfig", "GetInfo")
	} else {
		soapReq = fritz.CreateNewSoapData(*aI.Username, *aI.Password, *aI.Hostname, *aI.Port, "/upnp/control/wancommonifconfig1", "WANCommonInterfaceConfig", "GetCommonLinkProperties")
	}

	go fritz.DoSoapRequest(&soapReq, resps, errs, aI.Debug)

	res, err := fritz.ProcessSoapResponse(resps, errs, 1, *aI.Timeout)

	if err != nil {
		fmt.Printf("UNKNOWN - %s\n", err)
		return
	}

	var downstream float64

	if isDSL {
		soapResp := fritz.WANDSLInterfaceGetInfoResponse{}
		err = fritz.UnmarshalSoapResponse(&soapResp, res)

		if err != nil {
			panic(err)
		}

		ups, err := strconv.ParseFloat(soapResp.NewDownstreamCurrRate, 64)

		if err != nil {
			panic(err)
		}

		downstream = ups / float64(*aI.DivisorMax)
	} else {
		soapResp := fritz.WANCommonInterfaceCommonLinkPropertiesResponse{}
		err = fritz.UnmarshalSoapResponse(&soapResp, res)

		if err != nil {
			panic(err)
		}

		ups, err := strconv.ParseFloat(soapResp.NewLayer1DownstreamMaxBitRate, 64)

		if err != nil {
			panic(err)
		}

		downstream = ups / float64(*aI.DivisorMax)
	}

	perfData := perfdata.CreatePerformanceData("downstream_max", downstream, "")

	GlobalReturnCode = exitOk

	if thresholds.IsSet(aI.Warning) {
		perfData.SetWarning(*aI.Warning)

		if thresholds.CheckLower(*aI.Warning, downstream) {
			GlobalReturnCode = exitWarning
		}
	}

	if thresholds.IsSet(aI.Critical) {
		perfData.SetCritical(*aI.Critical)

		if thresholds.CheckLower(*aI.Critical, downstream) {
			GlobalReturnCode = exitCritical
		}
	}

	output := " - Max Downstream: " + fmt.Sprintf("%.2f", downstream) + " Mbit/s " + perfData.GetPerformanceDataAsString()

	switch GlobalReturnCode {
	case exitOk:
		fmt.Print("OK" + output + "\n")
	case exitWarning:
		fmt.Print("WARNING" + output + "\n")
	case exitCritical:
		fmt.Print("CRITICAL" + output + "\n")
	default:
		GlobalReturnCode = exitUnknown
		fmt.Print("UNKNWON - Not able to calculate maximum downstream\n")
	}
}

// CheckDownstreamCurrent checks the current used downstream
func CheckDownstreamCurrent(aI ArgumentInformation) {
	resps := make(chan []byte)
	errs := make(chan error)

	soapReq := fritz.CreateNewSoapData(*aI.Username, *aI.Password, *aI.Hostname, *aI.Port, "/upnp/control/wancommonifconfig1", "WANCommonInterfaceConfig", "X_AVM-DE_GetOnlineMonitor")
	soapReq.AddSoapDataVariable(fritz.CreateNewSoapVariable("NewSyncGroupIndex", "0"))
	go fritz.DoSoapRequest(&soapReq, resps, errs, aI.Debug)

	res, err := fritz.ProcessSoapResponse(resps, errs, 1, *aI.Timeout)

	if err != nil {
		fmt.Printf("UNKNOWN - %s\n", err)
		return
	}

	soapResp := fritz.WANCommonInterfaceOnlineMonitorResponse{}
	err = fritz.UnmarshalSoapResponse(&soapResp, res)

	downstreamWithHistory := strings.Split(soapResp.NewDSCurrentBPS, ",")

	downstream, err := strconv.ParseFloat(downstreamWithHistory[0], 64)

	if err != nil {
		panic(err)
	}

	downstream = downstream * 8 / float64(*aI.DivisorCurrent)
	perfData := perfdata.CreatePerformanceData("downstream_current", downstream, "")

	GlobalReturnCode = exitOk

	if thresholds.IsSet(aI.Warning) {
		perfData.SetWarning(*aI.Warning)

		if thresholds.CheckUpper(*aI.Warning, downstream) {
			GlobalReturnCode = exitWarning
		}
	}

	if thresholds.IsSet(aI.Critical) {
		perfData.SetCritical(*aI.Critical)

		if thresholds.CheckUpper(*aI.Critical, downstream) {
			GlobalReturnCode = exitCritical
		}
	}

	output := " - Current Downstream: " + fmt.Sprintf("%.2f", downstream) + " Mbit/s " + perfData.GetPerformanceDataAsString()

	switch GlobalReturnCode {
	case exitOk:
		fmt.Print("OK" + output + "\n")
	case exitWarning:
		fmt.Print("WARNING" + output + "\n")
	case exitCritical:
		fmt.Print("CRITICAL" + output + "\n")
	default:
		GlobalReturnCode = exitUnknown
		fmt.Print("UNKNWON - Not able to calculate current downstream\n")
	}
}

// CheckDownstreamUsage checks the current used downstream
func CheckDownstreamUsage(aI ArgumentInformation) {
	resps := make(chan []byte)
	errs := make(chan error)

	soapReq := fritz.CreateNewSoapData(*aI.Username, *aI.Password, *aI.Hostname, *aI.Port, "/upnp/control/wancommonifconfig1", "WANCommonInterfaceConfig", "X_AVM-DE_GetOnlineMonitor")
	soapReq.AddSoapDataVariable(fritz.CreateNewSoapVariable("NewSyncGroupIndex", "0"))
	go fritz.DoSoapRequest(&soapReq, resps, errs, aI.Debug)

	res, err := fritz.ProcessSoapResponse(resps, errs, 1, *aI.Timeout)

	if err != nil {
		fmt.Printf("UNKNOWN - %s\n", err)
		return
	}

	soapResp := fritz.WANCommonInterfaceOnlineMonitorResponse{}
	err = fritz.UnmarshalSoapResponse(&soapResp, res)

	downstreamWithHistory := strings.Split(soapResp.NewDSCurrentBPS, ",")
	downstreamCurrent, err := strconv.ParseFloat(downstreamWithHistory[0], 64)

	if err != nil {
		panic(err)
	}

	isDSL := false

	if strings.ToLower(*aI.Modelgroup) == "dsl" {
		isDSL = true
	}

	if isDSL {
		soapReq = fritz.CreateNewSoapData(*aI.Username, *aI.Password, *aI.Hostname, *aI.Port, "/upnp/control/wandslifconfig1", "WANDSLInterfaceConfig", "GetInfo")
	} else {
		soapReq = fritz.CreateNewSoapData(*aI.Username, *aI.Password, *aI.Hostname, *aI.Port, "/upnp/control/wancommonifconfig1", "WANCommonInterfaceConfig", "GetCommonLinkProperties")
	}

	go fritz.DoSoapRequest(&soapReq, resps, errs, aI.Debug)

	res, err = fritz.ProcessSoapResponse(resps, errs, 1, *aI.Timeout)

	if err != nil {
		fmt.Printf("UNKNOWN - %s\n", err)
		return
	}

	var downstreamMax float64

	if isDSL {
		soapResp := fritz.WANDSLInterfaceGetInfoResponse{}
		err = fritz.UnmarshalSoapResponse(&soapResp, res)

		if err != nil {
			panic(err)
		}

		ups, err := strconv.ParseFloat(soapResp.NewDownstreamCurrRate, 64)

		if err != nil {
			panic(err)
		}

		downstreamMax = ups / float64(*aI.DivisorMax)
	} else {
		soapResp := fritz.WANCommonInterfaceCommonLinkPropertiesResponse{}
		err = fritz.UnmarshalSoapResponse(&soapResp, res)

		if err != nil {
			panic(err)
		}

		ups, err := strconv.ParseFloat(soapResp.NewLayer1DownstreamMaxBitRate, 64)

		if err != nil {
			panic(err)
		}

		downstreamMax = ups / float64(*aI.DivisorMax)
	}

	downstreamCurrent = downstreamCurrent * 8 / float64(*aI.DivisorCurrent)

	if downstreamMax == 0 {
		fmt.Printf("UNKNOWN - Maximum Downstream is 0\n")
		return
	}

	downstreamUsage := 100 / downstreamMax * downstreamCurrent
	perfData := perfdata.CreatePerformanceData("downstream_usage", downstreamUsage, "")

	perfData.SetMinimum(0.0)
	perfData.SetMaximum(100.0)

	GlobalReturnCode = exitOk

	if thresholds.IsSet(aI.Warning) {
		perfData.SetWarning(*aI.Warning)

		if thresholds.CheckUpper(*aI.Warning, downstreamUsage) {
			GlobalReturnCode = exitWarning
		}
	}

	if thresholds.IsSet(aI.Critical) {
		perfData.SetCritical(*aI.Critical)

		if thresholds.CheckUpper(*aI.Critical, downstreamUsage) {
			GlobalReturnCode = exitCritical
		}
	}

	output := " - " + fmt.Sprintf("%.2f", downstreamUsage) + "% Downstream utilization (" + fmt.Sprintf("%.2f", downstreamCurrent) + " Mbit/s of " + fmt.Sprintf("%.2f", downstreamMax) + " Mbits) " + perfData.GetPerformanceDataAsString()

	switch GlobalReturnCode {
	case exitOk:
		fmt.Print("OK" + output + "\n")
	case exitWarning:
		fmt.Print("WARNING" + output + "\n")
	case exitCritical:
		fmt.Print("CRITICAL" + output + "\n")
	default:
		GlobalReturnCode = exitUnknown
		fmt.Print("UNKNWON - Not able to calculate downstream utilization\n")
	}
}
