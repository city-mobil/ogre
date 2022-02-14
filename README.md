# Local buffering proxy for Sentry.

Spoofs the response for the sentry client and adds the request to a local queue to be proxyed to the real sentry.
The service should be used to protect main Sentry instance from being DDoS'ed by clients.

## How to use:

```bash
ogre -h      
Usage of ./ogre:
  --debug
        debug mode (default false)
  --httpproxy.port string
        HTTPProxy port (default ":80")
  --httpproxy.server.error_threshold duration
        Threshold for error stats (default 3m0s)
  --httpproxy.server.idle_timeout duration
        HTTPProxy HTTP idle timeout (default 5s)
  --httpproxy.server.read_timeout duration
        HTTPProxy HTTP read timeout (default 1s)
  --logger_output string
        Use File or Stdout for output (default "file")
  --pprof string
        use with debug mode (default ":7001")
  --pusher.client.request_timeout duration
        Pusher HTTP client timeout (default 100ms)
  --pusher.drop_threshold duration
        Event drop threshold duration (default 20ms)
  --pusher.pool.unstoppable_workers int
        Pusher pool unstoppable workers count. (default 4)
  --pusher.queue_size int
        Pusher queue size (default 1000)
  --pusher.sentry.host string
        Pusher address of sentry hosts, use comma to separate hostnames
  --pusher.sentry.projects string
        Map with projects for sentry. Checkout readme for more info.
  --version
        version info
```

### Using multiple sentry instances
So far, two sentry instances are supported. To do this, in the `--pusher.sentry.host` parameter, list the sentry addresses separated by commas, necessarily with the scheme. For example, `--pusher.sentry.host http://host1.com:9000,https://host2.ru:8000`
In the `--pusher.sentry.projects` parameter, you must specify the correspondence of the project IDs in both sentries in the format `ID1:ID2,ID3:ID4`.

How it works:
Let's say we have a python project in sentry1. It has its own unique ID, which can be found at the URL `/settings/sentry/projects/{project_name}/keys/`. To write to sentry2 in the same project, we need to specify that ID2=ID1, where ID2 is the ID of the `python` project in the second sentry. You can specify multiple matches separated by commas.

Example:
```bash
./ogre --pusher.sentry.host http://localhost:9000,https://example.com:9000 --pusher.sentry.projects b1302c1848c343bdbec1a86e19288f0b:ae89dbc83adc43829e0b3097561c78ba,3cf5af72ad8b4a4baa5715b972e4d738:b4329234d48b44f68a8a4f3a07e3d9a7
```  
**Notice**: the order in which the project IDs are specified is important - ID1 should refer to host1, and ID2 to host2.

## Monitoring
### Status
```bash
curl -s  http://localhost:80/_status  | python -mjson.tool
{
    "author": "i.ivanov",
    "commit": "1e816509",
    "datetime": "2020-02-01T13:42:19+0300",
    "status": "OK"
}
```

### Prometheus metrics
```bash
curl http://localhost:80/metrics | egrep 'ogre|queue'
# HELP ogre_failed_put_to_queue_total
# TYPE ogre_failed_put_to_queue_total counter
ogre_failed_put_to_queue_total 41
# HELP ogre_info
# TYPE ogre_info counter
ogre_info{version=""} 1
# HELP ogre_put_to_queue_total
# TYPE ogre_put_to_queue_total counter
ogre_put_to_queue_total 5500
# HELP ogre_req_count
# TYPE ogre_req_count counter
ogre_req_count 2
# HELP ogre_sentry_failed_request_total
# TYPE ogre_sentry_failed_request_total counter
ogre_sentry_failed_request_total 1
# HELP ogre_sentry_request_total
# TYPE ogre_sentry_request_total counter
ogre_sentry_request_total 5499
# HELP queue_size_ratio
# TYPE queue_size_ratio gauge
queue_size_ratio -4.36
```

## Disclaimer

All information and source code are provided AS-IS, without express or implied warranties. 
Use of the source code or parts of it is at your sole discretion and risk. 
Citymobil LLC takes reasonable measures to ensure the relevance of the information posted in this repository, but it does not assume responsibility for maintaining or updating this repository or its parts outside the framework established by the company independently and without notifying third parties.


Вся информация и исходный код предоставляются в исходном виде, без явно выраженных или подразумеваемых гарантий. Использование исходного кода или его части осуществляются исключительно по вашему усмотрению и на ваш риск. Компания ООО "Ситимобил" принимает разумные меры для обеспечения актуальности информации, размещенной в данном репозитории, но она не принимает на себя ответственности за поддержку или актуализацию данного репозитория или его частей вне рамок, устанавливаемых компанией самостоятельно и без уведомления третьих лиц.