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

API ONDERSTEUNT JPEG, PNG EN (webp en avif nog niet)

**Gebruik community sdk**

OPENAI HEEFT GEEN OFFICIELE SDK



# 10 MB MAX VOOR ELKE FOTO, als file size te groot, dan error terugsturen. 



1. Water Supply Check
Route: /api/laundry/silver/v1/waterFeedAttachedToTap
Check: Water supply hose connected to tap
System Prompt: Looks for secure water connection, no leaks
Applies to: Washing machines, dishwashers, dryers with water connections

2. Drain Hose Check
Route: /api/laundry/silver/v1/drainHoseInDrain
Check: Drain hose connected to drain pipe
System Prompt: Looks for proper drain connection
Applies to: Washing machines, dishwashers, dryers with drain connections

3. Power Cord Check
Route: /api/laundry/silver/v1/powerCordInSocket
Check: Power cord plugged into socket
System Prompt: Looks for secure electrical connection
Applies to: All electrical appliances (washing machines, dryers, dishwashers, etc.)

4. Rinse Cycle Check
Route: /api/laundry/silver/v1/rinseCycleMachineIsOn
Check: Appliance running rinse cycle
System Prompt: Looks for active appliance operation
Applies to: Washing machines, dishwashers with rinse cycles

5. Shipping Bolts Check
Route: /api/laundry/silver/v1/shippingBoltsRemoved
Check: Transport bolts removed
System Prompt: Looks for removed shipping bolts
Applies to: All appliances with shipping bolts (washing machines, dryers, etc.)

6. Level Indicator Check
Route: /api/laundry/silver/v1/levelIndicatorPresent
Check: Spirit level present
System Prompt: Looks for leveling tool
Applies to: All appliances requiring leveling (washing machines, dryers, dishwashers, etc.)


## TO DO
# GOLD enpoint endpoints  (done)
# API KEY INVENTARISATIE & IMPLEMENTATIE
# Rate limits
# DATABASE IMPLEMENTEREN EN TESTEN
# DASHBOARD ONTWIKKELEN
# Dataset van Sathena nog testen

