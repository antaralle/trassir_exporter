## Мониторинг серверов видеонаблюдения TRASSIR

### Установка и настройка

Первым делом что бы экспортер работал необходимо включить на сервере за которым будем наблюдать веб сервер(порт 8080) и SDK что бы получить доступ к АПИ 

![trassir](https://i122.fastpic.org/big/2023/1024/7c/60d45d136787e004bdb3f644bb5f847c.png?md5=-Q8w9Zdr14t6jSjnesveWw&expires=1698138000)

В папке программы должен находиться файлик config.yaml - в yaml формате с логином паролем:
```yaml
username: testuser
password: 123
server: localhost
```

Далее просто запускаем exeшник и проверяем работу что всё нормально, зайдя на http://127.0.0.1:8080/metrics

### Описание собираемых метрик
![mectics](https://i122.fastpic.org/big/2023/1024/66/b50d907fdaea33a4d9452c82d4585f66.jpg?md5=q0tqtQF53yKWlfS-xOAo6g&expires=1698138000)

Подробнее можно почитать на сайте с описание SDK Трассира - https://www.dssl.ru/files/trassir/manual/ru/sdk-examples-health.html

### Настройка Prometheus

prometheus.yml
```yaml
  - job_name: 'trassir-metrics'
    scrape_interval: 1m
    static_configs:
    - targets:
      - '192.168.1.135:8080'
```
Тут всё стандартно, но обратите внимание на job_name, потом оно будет участвовать в дашборде Графаны и если решите поменять название, то не забудьте потом сменить и в Графане

Настройка алертов, alert.rules

```yaml
- name: trassir problems
  rules:
  - alert: trassir disks problem
    expr: disks{job="trassir-metrics"} == 0
    for: 1m
    labels:
      severity: critical

  - alert: trassir bd problem
    expr: database{job="trassir-metrics"} == 0
    for: 1m
    labels:
      severity: critical

  - alert: trassir network problem
    expr: network{job="trassir-metrics"} == 0
    for: 1m
    labels:
      severity: critical

  - alert: trassir automation problem
    expr: automation{job="trassir-metrics"} == 0
    for: 1m
    labels:
      severity: critical

```

### Дашборд Grafana
Выводит метрики которые были описаны выше. 

![grafana](https://i122.fastpic.org/big/2023/1024/5c/4741b090d80c4ac4a9d10fcbe618825c.jpg?md5=HsoQDFj5g4g5_V-GaXUmGw&expires=1698138000)

Сам дашборд в папке grafana