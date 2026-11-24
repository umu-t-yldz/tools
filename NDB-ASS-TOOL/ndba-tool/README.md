
## Derleme

```bash
cd ndba-tool
go build -o flightsim .
```

## Hızlı Başlangıç (dosya moduyla)

1) `c2.txt` dosyasını satır başına bir hedef olacak şekilde oluşturun (URL, domain veya IP):
```text
google.com
facebook.com
8.8.8.8
https://www.youtube.com/watch?v=2SmibCPWErs
```

2) `run` komutunu dosya modunda çalıştırıp sonucu JSON olarak yazın:
```bash
./flightsim run -file c2.txt -out result.json -timeout 10
```

- `-file`: Girdi dosyası (satır başına bir hedef)
- `-out`: Çıkış JSON dosya yolu (varsayılan: `result.json`)
- `-timeout`: Hedef başına zaman aşımı (saniye, varsayılan: 10)

### Çıktı şeması (result.json)

Her kayıt tek bir `status` alanı içerir:
```json
[
  { "target": "google.com", "category": "domain", "status": "connected" },
  { "target": "8.8.8.8",   "category": "ip",     "status": "connected" },
  { "target": "http://bad.example", "category": "url", "status": "not_connected", "error": "..." }
]
```

- `category`: `domain` | `ip` | `url`
- `status`: `connected` veya `not_connected`
- `error`: Sadece `not_connected` durumunda dolar

