package spoolfile

import (
	"testing"
	"fmt"
	"github.com/griesbacher/nagflux/helper"
)

var TestPerformanceData = []struct {
	input    string
	expected []PerformanceData
}{
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791000	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{},
			fields:           map[string]string{"value":"4.0"},
		}},
	}, {
		`DATATYPE::SERVICEPERFDATA	TIMET::1441791000	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4 'C:\ used %'=44%;89;94;0;100	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1`,
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{},
			fields:           map[string]string{"value":"4.0"},
		}, {
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: `'C:\ used %'`,
			unit:             "%",
			tags:             map[string]string{"warn-fill":"none", "crit-fill":"none"},
			fields:           map[string]string{"value":"44.0", "warn":"89.0", "crit":"94.0", "min":"0.0", "max":"100.0"},
		}},
	},
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791001	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4;2;10	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791001000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{"warn-fill":"none", "crit-fill":"none"},
			fields:           map[string]string{"value":"4.0", "warn":"2.0", "crit":"10.0"},
		}},
	},
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791002	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4;2;10;1;4	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791002000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{"warn-fill":"none", "crit-fill":"none"},
			fields:           map[string]string{"value":"4.0", "warn":"2.0", "crit":"10.0", "min":"1.0", "max":"4.0"},
		}},
	},
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791003	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4;2:4;8:10;1;4	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791003000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{"warn-fill":"outer", "crit-fill":"outer"},
			fields:           map[string]string{"value":"4.0", "warn-min":"2.0", "warn-max":"4.0", "crit-min":"8.0", "crit-max":"10.0", "min":"1.0", "max":"4.0"},
		}},
	},
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791004	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4;@2:4;@8:10;1;4	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791004000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{"warn-fill":"inner", "crit-fill":"inner"},
			fields:           map[string]string{"value":"4.0", "warn-min":"2.0", "warn-max":"4.0", "crit-min":"8.0", "crit-max":"10.0", "min":"1.0", "max":"4.0"},
		}},
	},
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791005	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4;2:;10:;1;4	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791005000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{"warn-fill":"none", "crit-fill":"none"},
			fields:           map[string]string{"value":"4.0", "warn":"2.0", "crit":"10.0", "min":"1.0", "max":"4.0"},
		}},
	},
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791006	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4;:2;:10;1;4	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791006000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{"warn-fill":"none", "crit-fill":"none"},
			fields:           map[string]string{"value":"4.0", "warn":"2.0", "crit":"10.0", "min":"1.0", "max":"4.0"},
		}},
	},
	{
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791007	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4;~:2;10:~;1;4	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791007000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{"warn-fill":"none", "crit-fill":"none"},
			fields:           map[string]string{"value":"4.0", "warn":"2.0", "crit":"10.0", "min":"1.0", "max":"4.0"},
		}},
	},
	{
		//test dot separated data
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791000	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::a used=4.5	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: "a used",
			unit:             "",
			tags:             map[string]string{},
			fields:           map[string]string{"value":"4.5"},
		}},
	}, {
		//test comma separated data
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791000	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::comma=4,5	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: "comma",
			unit:             "",
			tags:             map[string]string{},
			fields:           map[string]string{"value":"4.5"},
		}},
	},{
		//test tag
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791000	NAGFLUX:TAG::foo=bar	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::tag=4.5	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: "tag",
			unit:             "",
			tags:             map[string]string{"foo":"bar"},
			fields:           map[string]string{"value":"4.5"},
		}},
	},{
		//test empty tag
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791000	NAGFLUX:TAG::	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::tag=4.5	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: "tag",
			unit:             "",
			tags:             map[string]string{},
			fields:           map[string]string{"value":"4.5"},
		}},
	},{
		//test malformed tag
		"DATATYPE::SERVICEPERFDATA	TIMET::1441791000	NAGFLUX:TAG::$_SERVICENAGFLUX_TAG$	HOSTNAME::xxx	SERVICEDESC::range	SERVICEPERFDATA::tag=4.5	SERVICECHECKCOMMAND::check_ranges!-w 3: -c 4: -g :46 -l :48 SERVICESTATE::0	SERVICESTATETYPE::1",
		[]PerformanceData{{
			hostname:         "xxx",
			service:          "range",
			command:          "check_ranges",
			time:             "1441791000000",
			performanceLabel: "tag",
			unit:             "",
			tags:             map[string]string{},
			fields:           map[string]string{"value":"4.5"},
		}},
	},
}

func compareStringMap(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}
	return true
}

func comparePerformanceData(p1, p2 PerformanceData) (bool, string) {
	if p1.hostname != p2.hostname {
		return false, "hostname:" + p1.hostname + "!=" + p2.hostname
	}
	if p1.service != p2.service {
		return false, "service:" + p1.service + "!=" + p2.service
	}
	if p1.command != p2.command {
		return false, "command:" + p1.command + "!=" + p2.command
	}
	if p1.time != p2.time {
		return false, "time:" + p1.time + "!=" + p2.time
	}
	if p1.performanceLabel != p2.performanceLabel {
		return false, "performanceLabel:" + p1.performanceLabel + "!=" + p2.performanceLabel
	}
	if p1.unit != p2.unit {
		return false, "unit:" + p1.unit + "!=" + p2.unit
	}
	if !compareStringMap(p1.tags, p2.tags) {
		return false, "tags:" + fmt.Sprint(p1.tags) + "!=" + fmt.Sprint(p2.tags)
	}
	if !compareStringMap(p1.fields, p2.fields) {
		return false, "fields:" + fmt.Sprint(p1.fields) + "!=" + fmt.Sprint(p2.fields)
	}
	return true, "equal"
}

func TestNagiosSpoolfileWorker_PerformanceDataIterator(t *testing.T) {
	w := NewNagiosSpoolfileWorker(0, nil, nil, nil, 4096)
	for _, data := range (TestPerformanceData) {
		splittedPerformanceData := helper.StringToMap(data.input, "\t", "::")
		for singlePerfdata := range w.PerformanceDataIterator(splittedPerformanceData) {
			found := false
			for _, expectedPerfdata := range data.expected {
				equal, _ := comparePerformanceData(singlePerfdata, expectedPerfdata)
				if equal {
					found = true
					break
				}
			}
			if !found {
				t.Error("The expected perfdata was not found:", singlePerfdata, "\nRaw data:", data)
			}
		}
	}
}
