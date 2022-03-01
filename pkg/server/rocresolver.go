// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/miekg/dns"
	"github.com/onosproject/subscriber-dns/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// TODO: import definition from ROC package
type DeviceState struct {
	Attached     string `json:"attached"`
	Description  string `json:"description"`
	DeviceGroups string `json:"device_groups"`
	ID           string `json:"id"`
	IMEI         string `json:"imei"`
	IP           string `json:"ip"`
	Name         string `json:"name"`
	SimIccid     string `json:"sim_iccid"`
}

type RocResolver struct {
	baseUrl *url.URL // base URL to query device states
	client  *http.Client
}

func NewRocResolver(ctx context.Context, roc config.Roc, siteId string) (*RocResolver, error) {
	client := &http.Client{}
	if roc.OpenIDC != (config.OpenIDC{}) {
		cc := &clientcredentials.Config{
			TokenURL:  roc.OpenIDC.TokenUrl,
			ClientID:  "aether-roc-gui",
			Scopes:    []string{"openid", "profile", "email", "groups"},
			AuthStyle: oauth2.AuthStyleInHeader,
			EndpointParams: url.Values{
				"grant_type": {"password"},
				"username":   {roc.OpenIDC.Username},
				"password":   {roc.OpenIDC.Password},
			},
		}
		client = cc.Client(ctx)
	}

	baseUrl, _ := url.Parse(roc.Url)
	baseUrl.Path = path.Join(baseUrl.Path, siteId, "devices")

	return &RocResolver{
		baseUrl: baseUrl,
		client:  client,
	}, nil
}

func (r *RocResolver) HandleQuery(resp dns.ResponseWriter, req *dns.Msg) {
	log.Info("received query ", req.Question[0].String())

	m := new(dns.Msg)
	m.SetReply(req)
	m.Authoritative = true
	m.Compress = false
	m.RecursionAvailable = false

	switch req.Question[0].Qtype {
	case dns.TypeA:
		rr, rcode := r.deviceLookup(req.Question[0].Name)
		log.Debug("device lookup result: ", dns.RcodeToString[rcode])

		if rcode != dns.RcodeSuccess {
			m.SetRcode(req, rcode)
			log.Infof("sending answer ;%s \t%s", req.Question[0].Name, dns.RcodeToString[rcode])
		} else {
			m.Answer = append(m.Answer, rr)
			log.Info("sending answer ;", rr.String())
		}
	default:
		m.SetRcode(req, dns.RcodeNotImplemented)
		log.Warn("unsupported query type")
	}

	if err := resp.WriteMsg(m); err != nil {
		log.Error(err)
	}
}

func (r *RocResolver) deviceLookup(query string) (dns.RR, int) {
	reqUrl, _ := url.Parse(r.baseUrl.String())
	reqUrl.Path = path.Join(reqUrl.Path, strings.Split(query, ".")[0])

	log.Debug("requesting device info to ", reqUrl.String())

	resp, err := r.client.Get(reqUrl.String())
	if err != nil {
		log.Error(err)
		return nil, dns.RcodeServerFailure
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			log.Warnf("device %s does not exist", query)
			return nil, dns.RcodeNameError
		} else {
			log.Warnf("failed to get %s state(%s)", query, resp.Status)
			return nil, dns.RcodeServerFailure
		}
	}

	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return nil, dns.RcodeServerFailure
	}

	var respObj DeviceState
	if err := json.Unmarshal(respData, &respObj); err != nil {
		log.Error(err)
		return nil, dns.RcodeServerFailure
	}

	log.Debug("received response from ROC: ", respObj)

	if !validate(respObj) {
		log.Infof("device %s is not active", query)
		return nil, dns.RcodeNameError
	}

	rr, err := dns.NewRR(fmt.Sprintf("%s A %s", query, respObj.IP))
	if err != nil {
		log.Error(err)
		return nil, dns.RcodeServerFailure
	}

	return rr, dns.RcodeSuccess
}

func validate(d DeviceState) bool {
	return d.Attached == "!" && d.IP != ""
}
