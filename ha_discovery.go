package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/creativeprojects/go-homie"
)

// Home Assistant MQTT discovery.
//
// Home Assistant has no Homie integration, so without discovery every entity
// has to be written by hand in the Home Assistant config. With discovery the
// monitor publishes one retained config message per Homie property on
// <prefix>/<component>/hargassner/<node>_<property>/config. The configs only
// point at the existing Homie topics; those stay unchanged.

const homieDeviceID = "hargassner"

// haEntity describes one Homie property for Home Assistant.
type haEntity struct {
	Node     string
	Property string
	Name     MultiLanguageString
	Unit     string
	DataType homie.PropertyType
}

// haOptions are Home Assistant settings that cannot be derived from the
// property itself.
type haOptions struct {
	DeviceClass string
	StateClass  string
	Icon        string
	// LegacyUniqueID is the unique_id of the entity in the former hand-written
	// Home Assistant config. Home Assistant keeps entity ID and history only if
	// the discovered entity uses the same unique_id.
	LegacyUniqueID string
}

// haOverrides is keyed by "<node>/<property>".
var haOverrides = map[string]haOptions{
	"heizkreis1/vorlaufTemperatur":           {LegacyUniqueID: "hargassner_heizkreis1_vorlauftemperatur"},
	"heizkreis1/vorlaufSollTemperatur":       {LegacyUniqueID: "hargassner_heizkreis1_vorlaufsolltemperatur"},
	"heizkreis2/vorlaufTemperatur":           {LegacyUniqueID: "hargassner_heizkreis2_vorlauftemperatur"},
	"heizkreis2/vorlaufSollTemperatur":       {LegacyUniqueID: "hargassner_heizkreis2_vorlaufsolltemperatur"},
	"prozesswerte/stromAscheaustragung":      {LegacyUniqueID: "hargassner_strom_ascheaustragung"},
	"prozesswerte/stromEinschub":             {LegacyUniqueID: "hargassner_strom_einschub"},
	"prozesswerte/stromRaumaustragung":       {LegacyUniqueID: "hargassner_strom_raumaustragung"},
	"prozesswerte/aussenTemperaturAktuell":   {LegacyUniqueID: "hargassner_aussentemperatur_aktuell"},
	"prozesswerte/aussenTemperaturGemittelt": {LegacyUniqueID: "hargassner_aussentemperatur_gemittelt"},
	"prozesswerte/rauchgasTemperatur":        {LegacyUniqueID: "hargassner_rauchgastemperatur"},
	"prozesswerte/kesselTemperatur":          {LegacyUniqueID: "hargassner_kesseltemperatur"},
	"prozesswerte/boiler1Temperatur":         {LegacyUniqueID: "hargassner_boiler1_temperatur"},
	"prozesswerte/saugluftGeblaese":          {LegacyUniqueID: "hargassner_saugluftgeblaese"},
	"prozesswerte/primaerLuftGeblaese":       {LegacyUniqueID: "hargassner_primaerluftgeblaese"},
	"prozesswerte/o2InAbgas":                 {LegacyUniqueID: "hargassner_o2_in_abgas"},
	"prozesswerte/foerderMenge":              {LegacyUniqueID: "hargassner_foerdermenge"},
	"prozesswerte/meldung":                   {LegacyUniqueID: "unique_id__hargassner_meldung", Icon: "mdi:message-text-outline"},
	"kessel/AnzahlZuendungen":                {LegacyUniqueID: "hargassner_kessel_anzahl_zuendungen", StateClass: "total_increasing", Icon: "mdi:counter"},
	"kessel/DauerLetzteZuendung":             {LegacyUniqueID: "hargassner_kessel_dauer_letzte_zuendung", Icon: "mdi:timer-outline"},
	"kessel/DauerLetzterLeistungsbrand":      {LegacyUniqueID: "hargassner_kessel_dauer_letzter_leistungsbrand", Icon: "mdi:timer-outline"},
	"stoerung/lastActive":                    {LegacyUniqueID: "hargassner_stoerung_last_active", DeviceClass: "timestamp", Icon: "mdi:clock-outline"},
	"stoerung/nr":                            {LegacyUniqueID: "hargassner_stoerung_nr", Icon: "mdi:counter"},
	"stoerung/active":                        {LegacyUniqueID: "hargassner_stoerung_aktiv", DeviceClass: "problem"},
	"stoerung/text":                          {Icon: "mdi:alert-circle-outline"},
}

// haDeviceClassByUnit maps units to Home Assistant sensor device classes.
var haDeviceClassByUnit = map[string]string{
	"°C": "temperature",
	"A":  "current",
	"Pa": "pressure",
	"s":  "duration",
}

// haEntities holds all registered properties, keyed by "<node>/<property>".
// A map, because the records can be registered again (tests do that).
var haEntities = map[string]haEntity{}

func registerHAEntity(e haEntity) {
	haEntities[e.key()] = e
}

func (e haEntity) key() string {
	return e.Node + "/" + e.Property
}

func (e haEntity) component() string {
	if e.DataType == homie.TypeBoolean {
		return "binary_sensor"
	}
	return "sensor"
}

func (e haEntity) configTopic(prefix string) string {
	return fmt.Sprintf("%s/%s/%s/%s_%s/config", prefix, e.component(), homieDeviceID, e.Node, e.Property)
}

type haAvailability struct {
	Topic         string `json:"topic"`
	ValueTemplate string `json:"value_template"`
}

type haDevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SwVersion    string   `json:"sw_version"`
}

type haOrigin struct {
	Name string `json:"name"`
	Sw   string `json:"sw"`
	URL  string `json:"url"`
}

type haConfig struct {
	Name              string           `json:"name"`
	UniqueID          string           `json:"unique_id"`
	DefaultEntityID   string           `json:"default_entity_id,omitempty"`
	StateTopic        string           `json:"state_topic"`
	Availability      []haAvailability `json:"availability"`
	Device            haDevice         `json:"device"`
	Origin            haOrigin         `json:"origin"`
	UnitOfMeasurement string           `json:"unit_of_measurement,omitempty"`
	DeviceClass       string           `json:"device_class,omitempty"`
	StateClass        string           `json:"state_class,omitempty"`
	Icon              string           `json:"icon,omitempty"`
	PayloadOn         string           `json:"payload_on,omitempty"`
	PayloadOff        string           `json:"payload_off,omitempty"`
}

func (e haEntity) config(language string) haConfig {
	opts := haOverrides[e.key()]
	homieBase := "homie/" + homieDeviceID

	name := e.Name.EN
	if language == "de" {
		name = e.Name.DE
	}

	c := haConfig{
		Name:       name,
		UniqueID:   opts.LegacyUniqueID,
		StateTopic: homieBase + "/" + e.Node + "/" + e.Property,
		Availability: []haAvailability{{
			Topic:         homieBase + "/$state",
			ValueTemplate: "{{ 'online' if value == 'ready' else 'offline' }}",
		}},
		Device: haDevice{
			Identifiers:  []string{homieDeviceID},
			Name:         "Hargassner Heizung",
			Manufacturer: "Hargassner",
			Model:        "HSV",
			SwVersion:    version,
		},
		Origin: haOrigin{
			Name: "hargassner-monitor",
			Sw:   version,
			URL:  "https://github.com/rhierlmeier/hargassner-monitor",
		},
		DeviceClass: opts.DeviceClass,
		StateClass:  opts.StateClass,
		Icon:        opts.Icon,
	}
	if c.UniqueID == "" {
		c.UniqueID = homieDeviceID + "_" + e.Node + "_" + strings.ToLower(e.Property)
		c.DefaultEntityID = e.component() + "." + c.UniqueID
	}

	switch e.DataType {
	case homie.TypeBoolean:
		// Homie booleans are "true"/"false", Home Assistant expects ON/OFF.
		c.PayloadOn = "true"
		c.PayloadOff = "false"
	case homie.TypeInteger, homie.TypeFloat:
		c.UnitOfMeasurement = e.Unit
		if c.DeviceClass == "" {
			c.DeviceClass = haDeviceClassByUnit[e.Unit]
		}
		// Long-term statistics only for measured values; a number without
		// unit (error number) is a code, not a measurement.
		if c.StateClass == "" && e.Unit != "" {
			c.StateClass = "measurement"
		}
	}
	return c
}

// parseHADiscovery parses HARGASSNER_HA_DISCOVERY: empty disables discovery,
// "all" selects every property, otherwise a comma-separated list of
// "<node>/<property>". It returns nil if discovery is disabled.
func parseHADiscovery(value string) (map[string]bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	selected := map[string]bool{}
	if value == "all" {
		for key := range haEntities {
			selected[key] = true
		}
		return selected, nil
	}
	for _, key := range strings.Split(value, ",") {
		key = strings.TrimSpace(key)
		if _, ok := haEntities[key]; !ok {
			return nil, fmt.Errorf("unknown property %q (expected <node>/<property>)", key)
		}
		selected[key] = true
	}
	return selected, nil
}

// haDiscoveryMessages returns the retained messages to publish, keyed by
// topic. Properties that are not selected get an empty payload, which
// removes an entity that was discovered earlier.
func haDiscoveryMessages(selected map[string]bool, prefix, language string) (map[string]string, error) {
	messages := map[string]string{}
	for key, e := range haEntities {
		payload := ""
		if selected[key] {
			data, err := json.Marshal(e.config(language))
			if err != nil {
				return nil, fmt.Errorf("config for %s: %w", key, err)
			}
			payload = string(data)
		}
		messages[e.configTopic(prefix)] = payload
	}
	return messages, nil
}

// HA discovery settings, set in main.
var (
	haDiscoverySelected map[string]bool
	haDiscoveryPrefix   = "homeassistant"
	haLanguage          = "en"
)

func publishHADiscovery() {
	if haDiscoverySelected == nil {
		return
	}
	messages, err := haDiscoveryMessages(haDiscoverySelected, haDiscoveryPrefix, haLanguage)
	if err != nil {
		log.Printf("could not build Home Assistant discovery: %v", err)
		return
	}
	topics := make([]string, 0, len(messages))
	for topic := range messages {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	for _, topic := range topics {
		mqttClient.Publish(topic, 1, true, messages[topic])
	}
	log.Printf("Published Home Assistant discovery for %d of %d properties", len(haDiscoverySelected), len(messages))
}
