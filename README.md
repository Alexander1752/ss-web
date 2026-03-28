# Proiect ss-web - echipa LF4A

Acest proiect este o aplicație web care primește imagini de la dispozitive mobile prin MQTT, le procesează cu OCR și permite căutarea pe baza textului extras.

## Arhitectura Proiectului

```
ss-web/
├── client/          # Frontend React + TypeScript + Vite
├── server/          # Backend Go API
├── broker/          # Configurare MQTT Mosquitto
├── scripts/         # Scripturi utilitare
├── docs/            # Documentație (inclusiv ghid securitate)
├── uploads/         # Imagini încărcate
└── docker-compose.yml
```

## Tehnologii Folosite

| Component | Tehnologie |
|-----------|------------|
| Frontend | React + TypeScript + Vite + TailwindCSS |
| Backend | Go (Golang) |
| Bază de date | MongoDB |
| Broker MQTT | Eclipse Mosquitto |
| Containerizare | Docker Compose |
| Autentificare | JWT - TODO: de implementat (vezi `docs/AUTH_IMPLEMENTATION.md`) |
| Securitate | mTLS (Mutual TLS) - TODO: de implementat |

---

## Cerințe preliminare

1. **Docker** - pentru rularea containerelor
2. **Node.js** (v16+, recomandat v24.0.2) - pentru development frontend
3. **npm** sau **yarn** (recomandat yarn v1.22.22)
4. **Git** - pentru clonarea proiectului

---

## Pași pentru Setup

### Pasul 1: Clonarea Proiectului

```bash
git clone <repository-url>
cd ss-web
```

### Pasul 2: Configurarea Variabilelor de Mediu ---optional momentan


Verifică/creează fișierul `.env` în directorul rădăcină:
```bash
# .env
UID=501                               # User ID local (obține cu `id -u`)
GID=20                                # Group ID local (obține cu `id -g`)
MONGO_INITDB_ROOT_USERNAME=admin      # Username MongoDB
MONGO_INITDB_ROOT_PASSWORD=supersecret # Parolă MongoDB
JWT_SECRET=dev-secret                 # Secret pentru JWT
AWS_ACCESS_KEY=local-aws-access       # Opțional: pentru S3
AWS_SECRET_KEY=local-aws-secret       # Opțional: pentru S3
AWS_REGION=us-east-1                  # Opțional: pentru S3
S3_BUCKET_NAME=local-bucket           # Opțional: pentru S3
MQTT_HOST_IP=192.168.1.95             # IP-ul host-ului pentru MQTT
```

### Pasul 3: Pornirea Proiectului

**Metoda 1: Script automat (recomandat)**

```bash
./start.sh
```

Acest script va:
1. Instala dependențele client (yarn install)
2. Porni containerele Docker (API, MongoDB, MQTT Broker)
3. Porni serverul de development Vite

**Metoda 2: Manual**

```bash
# Terminal 1 - Pornește containerele Docker
docker compose up -d

# Terminal 2 - Pornește frontend-ul
cd client
yarn install
yarn dev:poll
```

### Pasul 5: Verificarea Funcționării

După pornire, aplicația va fi disponibilă la:

| Serviciu | URL/Port |
|----------|----------|
| Frontend (Vite) | http://localhost:5173 |
| Backend API | http://localhost:8080 |
| MongoDB | localhost:27019 |
| MQTT Broker (mTLS) | localhost:8883 |
| MQTT Broker (plain) | localhost:1883 |

---

## Oprirea Proiectului

```bash
# Oprește toate serviciile
./scripts/dev-stop.sh

# Sau manual
docker compose down
```

---

## Utilizare

### 1. Autentificare

- Accesează http://localhost:5173
- Login sau înregistrează un cont nou
- Autentificarea folosește JWT tokens

### 2. Vizualizare Fotografii

- Navighează la pagina Photos
- Vezi imaginile capturate de dispozitive
- Folosește căutarea pentru a găsi imagini după textul extras

### 3. Management Dispozitive

- Accesează pagina Devices
- Vezi toate dispozitivele conectate
- Schimbă modul între Normal și Live

### 4. Căutare Avansată

- Caută imagini după text extras (OCR)
- Filtrează după interval de date
- Filtrează după dispozitiv

---

## Scripturi Utile

| Script | Descriere |
|--------|-----------|
| `./start.sh` | Pornește întregul stack |
| `./scripts/dev-start.sh` | Script complet de pornire |
| `./scripts/dev-stop.sh` | Oprește toate serviciile |
| `./scripts/send_image.py` | Trimite imagini test prin MQTT |
| `./scripts/seed_data.py` | Populează baza de date cu date test |

---

## Debugging

### Verificare containere Docker

```bash
docker ps
docker logs go-api
docker logs broker
docker logs mongo-db
```

### Verificare conectivitate MQTT

```bash
python scripts/send_image.py
```

### Loguri Frontend

```bash
cat .dev-runtime/client.log
```

---

## Structura Codului

### Client (Frontend)

```
client/src/
├── components/       # Componente reutilizabile
│   ├── devicesCards/ # Card-uri pentru dispozitive
│   └── photosCards/  # Card-uri pentru fotografii
├── contexts/         # Context React (Auth)
├── pages/            # Paginile aplicației
│   ├── devicesPage/  # Management dispozitive
│   ├── homePage/     # Pagina principală
│   ├── loginPage/    # Autentificare
│   └── photosPage/   # Vizualizare fotografii
├── App.tsx           # Componenta principală
└── Router.tsx        # Configurare rute
```

### Server (Backend)

```
server/
├── main.go           # Entry point
├── domain/           # Modele de date
├── repository/       # Acces bază de date
├── routes/           # Endpoint-uri HTTP
├── broker/           # Client MQTT
└── utils/            # Utilități
```

---

---

## Scripturi pentru Încărcarea Imaginilor

### Script: `send_image.py` - Trimitere imagine individuală - la acest laborator vom trimite asa pozele, ulterior se vor trimite prin esp-cam si prin aplicatia android

Acest script permite trimiterea unei singure imagini prin MQTT către server.

**Utilizare:**

```bash
# Trimitere imagine generată automat (pentru testare)
python3 scripts/send_image.py

# Trimitere fișier specificat
python3 scripts/send_image.py /cale/catre/imagine.jpg
```

**Configurare Device ID:**
În fișierul `send_image.py`, poți modifica:
```python
DEVICE_ID = "python-sender-1"  # ID unic pentru dispozitiv
DEVICE_NAME = "Python Test Device"  # Nume vizibil în UI
```

**Topic-uri MQTT utilizate:**
- `register/{DEVICE_ID}` - pentru înregistrarea dispozitivului
- `ssproject/images/{DEVICE_ID}` - pentru trimiterea imaginilor

---

## Pagina de Statistici (Grafice)

Pagina de **Statistics** se află în meniul aplicației și afișează grafice despre documentele procesate.

**Acces:** După autentificare → meniul **Statistics**

### Grafice disponibile:

| Grafic | Ce arată |
|--------|----------|
| **Control Type Distribution** | Tipurile de controale medicale: Angajare, Periodic, Adaptare, Reluare, Supraveghere, Alte |
| **Medical Opinion Results** | Rezultatele avizelor medicale: APT, APT Condiționat, Inapt Temporar, Inapt |

### Carduri sumar:

- **Total Files** - numărul total de documente procesate
- **FIT (APT)** - numărul de avize APT (apt pentru muncă)
- **Periodic Checks** - numărul de controale periodice

### Funcționalități:

- **Filtrare pe date** - selectează interval de date (default: ultimele 30 zile)
- **Toggle Bar/Pie** - schimbă între grafic bar și grafic pie
- **Refresh** - reîncarcă datele

---

## Configurare IP MQTT pentru clienti

### Unde găsești IP-ul serverului MQTT:

**0. In pagina devices din aplicatia web (acolo apare si portul):**

```
Ip-ul serverului MQTT este afisat partea de sus a paginii devices
```

**1. În fișierul `.env` din rădăcina proiectului:**

```bash
MQTT_HOST_IP=192.168.1.95  # ← Acest IP se folosește în aplicația mobilă
```

**2. Aflat automat la pornirea serverului:**

Când rulezi `./start.sh`, scriptul afișează:
```
Detected HOST_IP: 192.168.1.95
```

**3. Aflat manual (pe macOS):**

```bash
# Obține IP-ul local
ipconfig getifaddr en0
# sau
ipconfig getifaddr en1
```

### Configurare în aplicația mobilă:

În aplicația mobilă Android/iOS, setează:

| Parametru | Valoare |
|-----------|---------|
| **MQTT Host** | IP-ul din `.env` (ex: `192.168.1.95`) |
| **MQTT Port (mTLS)** | `8883` |
| **MQTT Port (plain)** | `1883` |
| **Topic pentru imagini** | `ssproject/images/{DEVICE_ID}` |
| **Topic pentru înregistrare** | `register/{DEVICE_ID}` |

### Certificate necesare pentru mTLS:

> **Notă:** Securitatea mTLS nu este implementată implicit. Pentru a activa conexiunea securizată:
> 1. Urmați ghidul din [`docs/SECURITY_IMPLEMENTATION.md`](docs/SECURITY_IMPLEMENTATION.md)
> 2. Generați certificatele necesare în directorul `secrets/`

Pentru conexiunea securizată, aplicația mobilă are nevoie de:
- `ca.crt` - Certificate Authority
- Certificat client generat de aceeași CA

---


### Gestionarea și Ștergerea Imaginilor

Există două modalități de a șterge imaginile din aplicație:

**1. Ștergere Individuală:**
- Fiecare card de imagine are un buton de ștergere (🗑️) în colțul din dreapta-sus.
- Utilizați această opțiune pentru a elimina imagini specifice (ex: cele capturate greșit).

**2. Ștergere Totală (Resetare):**
- Butonul **Delete All** (roșu) șterge **toate** imaginile din baza de date.
- Această funcție este utilă pentru a curăța baza de date înainte de o nouă sesiune de testare sau demonstrativă.



---

## Ce să verifici în aplicație

### 1. Pagina Photos
- Verifică dacă imaginile încărcate apar corect
- Verifică textul extras automat (OCR)
- Testează funcția de căutare
- Testează butoanele **Capture**, **Start Live**, **Stop Live**
- Testează ștergerea pozelor (buton trash pe fiecare poză)

### 2. Pagina Devices
- Verifică dacă dispozitivele se înregistrează
- Testează schimbarea modului Normal/Live

### 3. Pagina Statistics
- Verifică graficele cu distribuția tipurilor de control
- Verifică graficele cu rezultatele avizelor medicale

---

## Task-uri Practice pentru Laborator

### Task 1: Testare trimitere imagine prin script (obligatoriu)

1. Pornește aplicația web (`./start.sh`)
2. Deschide un terminal și navighează în folderul proiectului
3. Rulează scriptul de trimitere imagini:
   ```bash
   python3 scripts/send_image.py /cale/catre/o/imagine.jpg
   ```
4. Accesează pagina **Photos** în browser și verifică că imaginea a apărut
5. Verifică textul extras prin OCR (dacă imaginea conține text)

**Livrabil:** Screenshot cu imaginea încărcată vizibilă în pagina Photos


### Task 2: Explorare și analiză statistici (opțional - pentru timp suplimentar)

1. Încarcă mai multe imagini folosind `send_image.py` (repetă de mai multe ori):
   ```bash
   python3 scripts/send_image.py /cale/catre/imagine1.jpg
   python3 scripts/send_image.py /cale/catre/imagine2.jpg
   ```
2. Accesează pagina **Statistics**
3. Modifică intervalul de date și observă cum se schimbă graficele
4. Schimbă tipul graficului între **Bar** și **Pie** pentru ambele categorii
5. Notează câte documente sunt în fiecare categorie de aviz medical

**Livrabil:** Screenshot cu pagina Statistics afișând ambele grafice

---

## Suport

Pentru probleme sau întrebări, consultați:
- Documentația MQTT: https://mqtt.org/
- Documentația proiectului în `client/README.md`




## Controale Camera ESP (în pagina Photos)

În partea de sus a paginii Photos există controale pentru camera ESP (pentru Lab-ul de ESP):

| Buton | Funcție |
|-------|---------|
| **Capture** (albastru) | Trimite o comandă către cameră să facă o singură poză și să o trimită |
| **Start Live** (verde) | Pornește modul live - camera trimite poze continuu până primește Stop |
| **Stop Live** (roșu) | Oprește modul live - camera încetează să mai trimită poze |

### Funcționarea modului Live:

1. Apasă **Start Live** - camera începe să trimită poze la interval regulat
2. Pozele apar automat în lista de fotografii
3. Când vrei să oprești, apasă **Stop Live**
4. Camera primește comanda și se oprește din trimitere
