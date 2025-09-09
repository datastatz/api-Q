# 📄 README.md (conceptversie)

**API-Q — Kwaliteitscontrole API**

API-Q is een service die monteurs ondersteunt bij kwaliteitscontrole van installaties (bijvoorbeeld witgoed).
Monteurs sturen een foto van hun installatie samen met een korte beschrijving van de criteria naar onze API.
De API beoordeelt de foto met behulp van AI en geeft een pass/fail resultaat met een korte feedback.

**🎯 Doel**

Tijdsbesparing: monteurs hoeven niet te wachten op handmatige controle.

Betrouwbaarheid: consistente beoordeling volgens dezelfde regels.

Schaalbaarheid: makkelijk uit te breiden naar andere installaties (bijv. zonnepanelen of airco’s).

**🔀 Flow**

Client → [multipart] → Jouw API → [base64] → OpenAI API

**🖼️ IMAGE ONDERSTEUNING**

API ONDERSTEUNT JPEG, PNG EN WEBP

**Gebruik community sdk**

OPENAI HEEFT GEEN OFFICIELE SDK

**🌐 Talen**
195 talen support 

# WEBP NOT SUPPORTED (KAN AAN MODEL LIGGEN OF AAN MN CODE MOET DAT NOG BEKIJKEN)

# 10 MB MAX VOOR ALLE FOTOS SAMEN, MOET WSS NAAR 30 MB GAAN (NOG CHECKEN)

# DATABASE SCHEMA 

-- Bedrijven tabel
companies:
- id (primary key)
- company_name
- contact_email
- created_at
- is_active

-- API keys gekoppeld aan bedrijven
api_keys:
- id (primary key)
- company_id (foreign key)
- api_key (unique)
- created_at
- is_active

-- Request logs voor analytics
request_logs:
- id (primary key)
- company_id (foreign key)
- api_key (foreign key)
- timestamp
- photo_count
- language
- cost_per_request (voor berekening)

# Data Flow

1. Bedrijf doet request → Go API → Database log → Response
2. Bedrijf → Next.js login → Go API check → JWT token → Dashboard
3. Dashboard data → Next.js → Go API → Database → Response

# Wat wij zien in de dashboard

Bedrijf A: 150 requests deze maand, €75
Bedrijf B: 89 requests deze maand, €44.50
Bedrijf C: 203 requests deze maand, €101.50
TOTAAL: 442 requests, €221

# Wat bedrijf X ziet in haar dashboard

Jouw requests deze maand: 150
Jouw kosten deze maand: €75

# Request flow

1. Request komt binnen:

POST /quality-check
Headers: X-API-Key: ak_abc123def456...

2. Middleware controleert:
Is er een X-API-Key header?
Bestaat de API key in de database?
Is de API key actief?

3. Als geldig:
Request gaat door naar je endpoint
Je endpoint krijgt de request

4. Als ongeldig:
Error response wordt teruggestuurd
Request stopt hier

# Request logic

Elke request wordt gelogd, ook foutieve:
✅ Geldige request → Foto geanalyseerd → €0.20 in rekening gebracht
❌ Foutieve request (geen foto, verkeerde format, etc.) → Geen analyse → €0.20 in rekening gebracht

Je rekent voor API gebruik, niet voor succesvolle analyses
Net zoals andere API providers (OpenAI, AWS, etc.)

**// Admin credentials (in productie: in database of environment)
const ADMIN_USERNAME = "admin"
const ADMIN_PASSWORD = "admin123"     // Verander dit!
const JWT_SECRET = "stellarisdebeste" // Verander dit!**

# Analytics GO fixen lijkt op bullshit wat er staat