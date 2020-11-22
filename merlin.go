package merlin

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"time"
)

var userAgent string = "rcn-merlin-exporter (+https://github.com/kleinpa/rcn-merlin-exporter)"

type Client struct {
	BaseURL string
}

func extractNumber(s string) float64 {
	re := regexp.MustCompile(`(-?\d+(?:\.\d+)?)`)

	f, err := strconv.ParseFloat(re.FindString(s), 64)
	if err != nil {
		return 0.
	}
	return f
}

type ofdmData struct {
	ActiveSubCarriers     int32 `json:"Active_SubCarriers"`
	PLC                   float64
	PercentileBelowThresh int32                `json:"PercentileBelowThresh"`
	PilotCount            int32                `json:"Pilot_Count"`
	SubcarrierSpacing     int32                `json:"Subcarrier_Spacing"`
	Downstream            []ofdmDownstreamData `json:"Downstream"`
}
type ofdmDownstreamData struct {
	DownstreamPwr    float64
	ChannelFrequency float64
}

func (x *ofdmData) UnmarshalJSON(bs []byte) error {
	type alias ofdmData
	aux := &struct {
		PLC string `json:"PLC"`
		*alias
	}{
		alias: (*alias)(x),
	}
	if err := json.Unmarshal(bs, &aux); err != nil {
		return err
	}
	x.PLC = extractNumber(aux.PLC)

	// It seems like the first entry is an out-of-order duplicate,
	// maybe it is the PLC, aynyway remove it for now.
	x.Downstream = x.Downstream[1:]

	return nil
}
func (x *ofdmDownstreamData) UnmarshalJSON(bs []byte) error {
	aux := &struct {
		DownstreamPwr    string `json:"Downstream Pwr"`
		ChannelFrequency string `json:"Channel Frequency"`
	}{}
	if err := json.Unmarshal(bs, &aux); err != nil {
		return err
	}
	x.DownstreamPwr = extractNumber(aux.DownstreamPwr)
	x.ChannelFrequency = extractNumber(aux.ChannelFrequency)
	return nil
}
func (mer *Client) GetOfdmData() (ofdmData, error) {
	data := ofdmData{}
	err := mer.RequestJSON("merlin/modem_ofdm.cgi", &data)
	if err != nil {
		return data, err
	}
	return data, err
}

type mipData struct {
	CMTS            string `json:"CMTS"`
	ClientIP        string `json:"ClientIP"`
	DHCPServer      string `json:"Found_ON_DHCPserver"`
	Cf              string `json:"cf"`
	CpeMAC          string `json:"cpeMAC"`
	DownstreamSpeed string `json:"downstreamSpeed"`
	EncodedMac      string `json:"encodedmac"`
	Modem           string `json:"modem"`
	ModemIP         string `json:"modemIP"`
	UpstreamSpeed   string `json:"upstreamSpeed"`
}

func (mer *Client) GetMipData() (mipData, error) {
	// TODO: how is /lookup_mip_merlin-new.cgi different?
	data := mipData{}
	err := mer.RequestJSON("lookup_mip_merlin.cgi", &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

type configData struct {
	Client  string `json:"client"`
	Server  string `json:"server"`
	Company string `json:"company"`
}

func (mer *Client) GatherConfig() (configData, error) {
	data := configData{}
	err := mer.RequestJSON("merlin/CONFIG_json.cgi", &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func (mer *Client) RequestJSON(p string, v interface{}) error {
	client := http.Client{
		Timeout: time.Second * 20, // Timeout after 2 seconds
	}

	u, err := url.Parse(mer.BaseURL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, p)

	log.Printf("GET %s\n", u)
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", userAgent)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(v); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// XHR when browsing /merlin/
// /merlin/modem_ofdm.cgi?ip=<client-ip>&MAC=<modem-mac-no-colons>&cmtsindex=undefined&CMTS=smr-cbr1.sbo-smr.ma.cable.rcn.net
// /merlin/rfmodem_ds.cgi?ip=<client-ip>
// /merlin/rfmodem_us_Ver2.cgi?ip=<client-ip>&cmts=smr-cbr1.sbo-smr.ma.cable.rcn.net&modem=<modem-mac>
// /lookup_mip_merlin-new.cgi
// /merlin/CmCapab_PM.cgi?arg=<modem-mac-no-colons>
// /merlin/CONFIG_json.cgi
// /merlin/dhcp2.cgi
// /merlin/FFT.cgi?JSON_Taps=0,0,0.000976610233781518,-0.000976610233781518,-0.0024415255844538,0.000976610233781518,0.0024415255844538,-0.000488305116890759,-0.00341813581823531,0.0024415255844538,0.00292983070134455,-0.00292983070134455,-0.00732457675336139,0.00830118698714291,0.0151374586236135,-0.0229503404938657,0.998583964041603,0,-0.0126959330391597,0.0273450865458825,-0.0175789842080673,0.00390644093512607,-0.0195322046756304,-0.0112310176884875,-0.00146491535067228,0.00292983070134455,-0.00585966140268911,0.00195322046756304,0,0.00146491535067228,0.00195322046756304,0.00146491535067228,0.000976610233781518,-0.000976610233781518,0,0,-0.00390644093512607,-0.00146491535067228,-0.000976610233781518,0.00146491535067228,-0.00390644093512607,-0.000976610233781518,0,-0.00146491535067228,0.00488305116890759,0.000976610233781518,0.00292983070134455,-0.00195322046756304,0.00146491535067228,-0.00488305116890759
// /merlin/FFT.cgi?JSON_Taps=0,0,0.000976660536461726,-0.000976660536461726,-0.000976660536461726,0.00146499080469259,0,-0.00341831187761604,-0.000488330268230863,0.00878994482815553,0.000976660536461726,-0.0122082567057716,-0.00195332107292345,0.0224631923386197,-0.00830161455992467,-0.0620179440653196,0.995705416922729,0,0.0239281831433123,-0.053227999237164,-0.00830161455992467,-0.00292998160938518,0.0039066421458469,0.00732495402346294,0.00976660536461726,-0.0122082567057716,0.000976660536461726,0.00195332107292345,0.00292998160938518,-0.000976660536461726,0,-0.000976660536461726,0.00146499080469259,-0.0039066421458469,-0.00341831187761604,-0.00244165134115431,-0.000488330268230863,-0.000976660536461726,0.000976660536461726,0,-0.000488330268230863,-0.00244165134115431,-0.00244165134115431,-0.0039066421458469,0.00244165134115431,0.00439497241407776,0,-0.000976660536461726,0,0.000488330268230863
// /merlin/FFT.cgi?JSON_Taps=0,0,0.00292994422952663,-0.00439491634428995,-0.00488324038254439,0.00683653653556214,0.00683653653556214,-0.0112314528798521,-0.00927815672683433,0.0161146932623965,0.0107431288415976,-0.0288111182570119,-0.0166030173006509,0.0610405047818048,-0.0297877663335208,-0.183609838383669,0.969811539973315,0,0.040042571136864,-0.124522629754882,-0.0209979336449409,0.0332060346013018,-0.00732486057381658,-0.0195329615301775,0.00146497211476332,0,0,-0.00976648076508877,-0.00146497211476332,0.00292994422952663,0.00195329615301775,-0.00341826826778107,0.000488324038254439,0.00146497211476332,0.00341826826778107,0.00292994422952663,-0.000976648076508877,0.000976648076508877,0.00292994422952663,-0.000488324038254439,0.000976648076508877,0.00341826826778107,-0.00537156442079882,0,-0.000976648076508877,0.00244162019127219,0.00195329615301775,-0.000976648076508877,0.000976648076508877,0.00146497211476332
// /merlin/findMac.cgi
// /merlin/gateways/get_eero_id_by_IP.cgi
// /merlin/impaired_channels.cgi?cmts=smr-cbr1.sbo-smr.ma.cable.rcn.net&modem=<modem-mac>&modemidx=271&chassis_model=CBR8
// /merlin/insertToAMD.cgi
// /merlin/macmap_Ver2.cgi?MAC=<modem-mac-no-colons>
// /merlin/macmap-primaryDs.cgi?cmts=smr-cbr1.sbo-smr.ma.cable.rcn.net&MAC=<modem-mac>
// /merlin/macmap-snrAlarms.cgi?mac=<modem-mac-no-colons>
// /merlin/macmap-status.cgi?mac=<modem-mac-no-colons>&INT=1001
// /merlin/mdmArp.cgi?ip=<client-ip>
// /merlin/ModemModel.cgi?MAC=<modem-mac-no-colons>
// /merlin/pre_eq_Ver2.cgi?modem=<client-ip>
// /merlin/rfmodem_misc.cgi?ip=<client-ip>&MAC=<modem-mac-no-colons>
// /merlin/Widget/ping.cgi?ip=<client-ip>&proto=tcp

// XHR when browsing /
// /lookup_mip_merlin.cgi
// /rfmodem_info.cgi (NOTE: this includes more data than /merlin/rfmodem_info.cgi and )

// All requests have cache buster param like _=1605982015720
