import argparse
import json
import os
import re
import sys

try:
    import urllib2
    from cStringIO import StringIO

    v = 2
except ImportError:
    import urllib.request, urllib.error, urllib.parse
    from io import StringIO

    v = 3


def handle_args():
    parser = argparse.ArgumentParser(description='Dumps Tables from InfluxDB and writes them to files')
    parser.add_argument('--url',
                        dest='url',
                        default='http://127.0.0.1:8086/query?db=mydb',
                        help='URL to the InfluxDB with username, password... Default: http://127.0.0.1:8086/query?db=mydb',
                        type=str, )
    parser.add_argument('--file',
                        dest='file',
                        help='File with tablenames, one tablename per line',
                        default='',
                        type=str, )
    parser.add_argument('tablenames',
                        type=str,
                        nargs='*',
                        help='List of tabelnames')
    parser.add_argument('--target',
                        dest='target',
                        help='Target folder. Default: dump',
                        default='dump',
                        type=str, )
    parser.add_argument('--fieldSeparator',
                        dest='separator',
                        help='The fieldSeparator of nagflux genericfile format. Default: &',
                        default='&',
                        type=str, )
    parser.add_argument('--hostcheckAlias',
                        dest='alias',
                        help='The fictional name for an hostcheck. Default: hostcheck',
                        default='hostcheck',
                        type=str, )
    return parser.parse_args()


def create_target_folder(target):
    if not os.path.exists(target):
        os.makedirs(target)


def build_tablename_list(file_name, tables):
    if file_name != '' and os.path.isfile(file_name) and os.access(file_name, os.R_OK):
        tables.extend([line.rstrip('\n') for line in open(file_name)])
    return tables


def gen_structur(tables):
    messages = {}
    metrics = {}
    for t in tables:
        match = re.search(table_regex, t)
        if match:
            service = match.group(2)
            if match.group(2) == "":
                service = args.alias
            if match.group(1) in metrics:
                if service in metrics[match.group(1)]:
                    if match.group(4) in metrics[match.group(1)][service]:
                        metrics[match.group(1)][service][match.group(4)][match.group(5)] = t
                    else:
                        # new perflabel
                        metrics[match.group(1)][service][match.group(4)] = {match.group(5): t}
                else:
                    # new service
                    metrics[match.group(1)][service] = {match.group(4): {match.group(5): t},
                                                        command_const: match.group(3)}

            else:
                # new host
                metrics[match.group(1)] = {
                    service: {match.group(4): {match.group(5): t}, command_const: match.group(3)}}
        else:
            t_match = re.search(message_regex, t)
            if t_match:
                service = t_match.group(2)
                if t_match.group(2) == "":
                    service = args.alias
                if t_match.group(1) in messages:
                    messages[t_match.group(1)][service] = t
                else:
                    messages[t_match.group(1)] = {service: t}
    return metrics, messages


def raise_http_error(code, table):
    raise Exception('Got HTTP: %s for table: %s' % (code, table))


def query_data_for_table(url, table):
    table = table.replace("\\", "\\\\")
    url += '&format=json&epoch=ms&q='
    if v == 2:
        url += urllib2.quote('SELECT * FROM "%s"' % table)
        try:
            response = urllib2.urlopen(urllib2.Request(url))
        except urllib2.HTTPError as error:
            raise_http_error(error.code, table)
    elif v == 3:
        url += urllib.parse.quote('SELECT * FROM "%s"' % table)
        try:
            response = urllib.request.urlopen(urllib.request.Request(url))
        except urllib.error.HTTPError as error:
            raise_http_error(error.code, table)
    if response.code != 200:
        raise_http_error(error.code, table)
    if v == 2:
        json_object = json.loads(response.read())
    elif v == 3:
        json_object = json.loads(response.read().decode('utf8'))
    return json_object

def castString(string):
    if not isinstance(string, str):
        if v == 2:
            if isinstance(string, unicode):
                string = string.encode('utf-8')
            else:
                string = str(string)
        elif v == 3:
            string = str(string)
    return string

def escape_for_influxdb(string):
    string = castString(string)
    return string.replace(' ', '\ ').replace(',', '\,')


def parse_object(json_object):
    if len(json_object['results']) == 0 or len(json_object['results'][0]) == 0:
        print("EMPTY RESULT!")
        return ""
    json_object = json_object['results'][0]['series'][0]
    tags = []
    for index, v in enumerate(json_object['columns']):
        if v != 'value' and v != 'time':
            tags.append((index, escape_for_influxdb(v)))
    time_index = json_object['columns'].index("time")
    value_index = json_object['columns'].index("value")
    return json_object, tags, time_index, value_index


def parse_message(json_object, tags, time_index, value_index, host, service):
    data = StringIO()
    data.write('t_host')
    data.write(args.separator)
    data.write('t_service')
    data.write(args.separator)
    for tag in tags:
        data.write('t_')
        data.write(tag[1])
        data.write(escape_for_influxdb(args.separator))
    data.write('f_message')
    data.write(args.separator)
    data.write('time')
    data.write(args.separator)
    data.write('table')
    data.write('\n')
    for value in json_object['values']:
        data.write(escape_for_influxdb(host))
        data.write(args.separator)
        data.write(escape_for_influxdb(service))
        data.write(args.separator)
        for tag in tags:
            data.write(escape_for_influxdb(value[tag[0]]))
            data.write(args.separator)
        data.write(escape_InfluxDB_string(value[value_index]))
        data.write(args.separator)
        data.write(str(value[time_index]))
        data.write(args.separator)
        data.write('messages')
        data.write('\n')
    return data.getvalue()


def escape_InfluxDB_string(string):
    return '"""' + castString(string).replace('"', '\\"') + '"""'


def dump_messages(messages):
    i = 0
    for host in messages:
        for serivce in messages[host]:
            result = query_data_for_table(args.url, messages[host][serivce])
            if result['results'] and len(result['results'][0]) > 0:
                json_object, tags, time_index, value_index = parse_object(result)
                csv = parse_message(json_object, tags, time_index, value_index, host, serivce)
                write_data_to_file(csv, args.target, str(i) + ".txt")
                i += 1
    return i


def parse_perf_type(data, perfType, json_object, tags, time_index, value_index):
    if perfType == "value":
        perfType = "f_value"
    for value in json_object['values']:
        time = value[time_index]
        if time not in data:
            data[time] = {}
        if perfType != "warn" and perfType != "crit":
            # value,min,max
            data[time][perfType] = escape_for_influxdb(value[value_index])
            for tag in tags:
                if tag[1] == 'unit':
                    data[time]['t_unit'] = escape_for_influxdb(value[tag[0]])
                elif tag[1] == 'downtime':
                    data[time]['t_downtime'] = 'true'
        elif perfType == "warn" or perfType == "crit":
            for tag in tags:
                if tag[1] == 'type':
                    if value[tag[0]] == 'normal':
                        data[time]["f_" + perfType] = value[value_index]
                    elif value[tag[0]] == 'min':
                        data[time]["f_" + perfType + '-min'] = value[value_index]
                    elif value[tag[0]] == 'max':
                        data[time]["f_" + perfType + '-max'] = value[value_index]
                elif tag[1] == 'fill':
                    data[time]["t_" + perfType + "-fill"] = escape_for_influxdb(value[tag[0]])
        else:
            print("!!!Unexpected type:" + perfType)


def gen_metrics_string(m_data, host, service, command, perfLabel):
    out = StringIO()
    fix_tags = ['t_host', 't_service', 't_command', 't_performanceLabel', 'table']
    fix_values = [host, service, command, perfLabel, 'metrics']
    ava_tags = ['f_value', 'f_min', 'f_max',
                'f_warn', 'f_warn-min', 'f_warn_max', 'f_crit', 'f_crit-min', 'f_crit-max',
                't_warn-fill', 't_crit-fill',
                't_unit']
    for s in fix_tags:
        out.write(s)
        out.write(args.separator)
    for s in ava_tags:
        out.write(s)
        out.write(args.separator)
    out.write('time')
    out.write('\n')
    for t in m_data:
        for v in fix_values:
            out.write(escape_for_influxdb(v))
            out.write(args.separator)
        for s in ava_tags:
            if s in m_data[t]:
                out.write(escape_for_influxdb(str(m_data[t][s])))
            out.write(args.separator)
        out.write(str(t))
        out.write('\n')
    return out.getvalue()


def dump_metrics(metrics):
    i = files_written
    lenHosts = len(metrics)
    h = 0
    for host in metrics:
        h += 1
        lenServices = len(metrics[host])
        s = 0
        for service in metrics[host]:
            s += 1
            print("Host: "+str(h)+" / "+str(lenHosts)+" Service: "+str(s)+" / "+str(lenServices))
            sys.stdout.flush()
            command = metrics[host][service][command_const]
            for perfLabel in metrics[host][service]:
                data = {}
                if perfLabel == command_const:
                    continue
                for perfType in metrics[host][service][perfLabel]:
                    result = query_data_for_table(args.url, metrics[host][service][perfLabel][perfType])
                    if result['results'] and len(result['results'][0]) > 0:
                        json_object, tags, time_index, value_index = parse_object(result)
                        parse_perf_type(data, perfType, json_object, tags, time_index, value_index)
                out = gen_metrics_string(data, host, service, command, perfLabel)
                write_data_to_file(out, args.target, str(i) + ".txt")
                i += 1
                print("Written:", host, service, command, perfLabel)


def write_data_to_file(data, target, filename):
    f = open(os.path.join(target, filename), 'w')
    f.write(data + '\n')
    f.close()


if __name__ == '__main__':
    args = handle_args()
    table_regex = '^(.*?)' + args.separator + '(.*?)' + args.separator + '(.*?)' + args.separator + '(.*?)' + args.separator + '(.*?)$'
    message_regex = '^(.*?)' + args.separator + '(.*?)' + args.separator + 'messages$'
    command_const = args.separator + 'command' + args.separator
    create_target_folder(args.target)
    table_names = build_tablename_list(args.file, args.tablenames)
    metrics, messages = gen_structur(table_names)
    print("Dumping messages...")
    files_written = 0
    files_written += dump_messages(messages)
    print("Dumping metrics...")
    dump_metrics(metrics)
