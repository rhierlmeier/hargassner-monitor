# Arbeitspaket: Dauer von Zündung und Leistungsbrand berechnen

Die Hargassner Heizung sendet über die Z-Records Ereignisse über Beginn und Ende von Zündung und Leistungsbrand. 
Diese Ereignisse können verwendet werden, um die Dauer von Zündung und Leistungsbrand zu berechnen. 
Die Berechnung erfolgt durch die Abfrage der Zeitstempel der Ereignisse und die Berechnung der Dauer in Sekunden oder Minuten.

Auszug aus den Logs:

```aiignore
026/02/14 13:21:44 Handling Z record: fields:[z|14:10:40|Kessel|Zündung] <-- Hier beginnt die Zündung
2026/02/14 13:21:45 Handling Z record: fields:[z|14:10:40|Kessel|Zündung|Start]
2026/02/14 13:23:25 Handling Z record: fields:[z|14:12:20|Kessel|Zündung|Einschub]
2026/02/14 13:26:25 Handling Z record: fields:[z|14:15:20|Kessel|Zündung|Pause]
2026/02/14 13:28:25 Handling Z record: fields:[z|14:17:20|Kessel|Zündung|Reduziert]
2026/02/14 13:31:25 Handling Z record: fields:[z|14:20:20|Kessel|Leistungsbrand] <-- Hier beginnt der Leistungsbrand
2026/02/14 17:01:22 Handling Z record: fields:[z|17:50:18|Kessel|Entaschung|Start]
2026/02/14 17:01:22 Handling Z record: fields:[z|17:50:18|Kessel|Entaschung|Gebläse]
2026/02/14 17:11:03 Handling Z record: fields:[z|17:59:58|Kessel|Entaschung|Rost]
2026/02/14 17:11:37 Handling Z record: fields:[z|18:00:32|Kessel|Aus] <-- Leistungsbrand endet
```

Berechne die Dauer von Zündung und Leistungsbrand in Sekunden und publiziere beide Werte über MQTT.

Beide Werte werden unter dem Homie Node `Kessel` publiziert.
* Dauer Leistungsbrand wird unter der Home Property `DauerLetzterLeistungsbrand` publiziert.
* Dauer Zündung wird unter der Home Property `DauerLetzteZuendung` publiziert.

Summiere weiterhin die Anzahl der Zündvorgänge auf und publiziere diese unter der Home Property `AnzahlZuendungen`.



