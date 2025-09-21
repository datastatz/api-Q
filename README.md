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

API ONDERSTEUNT JPEG, PNG EN (webp nog niet)

**Gebruik community sdk**

OPENAI HEEFT GEEN OFFICIELE SDK



# 10 MB MAX VOOR ELKE FOTO




