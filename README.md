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
2. Admin/Company login → Next.js → Go API → JWT token
3. Dashboard data → Next.js → Go API → Database → Response