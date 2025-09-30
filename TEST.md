TEST APPROVED: ✅ 
TEST FAILED: ❌

# Testing /api/laundry/silver/v1/waterFeedAttachedToTap 

✅ 
curl -X POST http://localhost:8080/api/laundry/silver/v1/waterFeedAttachedToTap \
  -F "photo=@cmd/api/waterFeedAttachedToTap/water.jpg"

feedback: Resultaat was inderdaad fail. Prima!

✅ 
curl -X POST http://localhost:8080/api/laundry/silver/v1/waterFeedAttachedToTap \
  -F "photo=@cmd/api/waterFeedAttachedToTap/water2.png"

feedback: Resultaat was inderdaad fail. Prima!

✅ 
curl -X POST http://localhost:8080/api/laundry/silver/v1/waterFeedAttachedToTap \
  -F "photo=@cmd/api/waterFeedAttachedToTap/water3.png"

feedback: Resultaat was pass de eerste keer maar daarna niet. Ik kreeg vervolgens twee keer fail. 
Ik krijg nu PASS nadat de systemprompt is aangepast. Lijkt nu gefixt na de aanpassing in de system prompt.


# /api/laundry/silver/v1/drainHoseInDrain

✅ 
curl -X POST http://localhost:8080/api/laundry/silver/v1/drainHoseInDrain \
  -F "photo=@cmd/api/drainHoseInDrain/drain1.png"

feedback: resultaat was pass dus prima! 3x getest. 3x fail. 

✅ 
curl -X POST http://localhost:8080/api/laundry/silver/v1/drainHoseInDrain \
  -F "photo=@cmd/api/drainHoseInDrain/drain2.png"

feedback: system prompt aangepast nu wel gewoon PASS, dus prima 

**Belangrijke vraag is dan wel of het voldoende is om alleen slang in de buis te hebben? **

# /api/laundry/silver/v1/powerCordInSocket

✅ 
curl -X POST http://localhost:8080/api/laundry/silver/v1/powerCordInSocket \
  -F "photo=@cmd/api/powerCordInSocket/power1.png"

feedback: Pass, correct!

✅ 
curl -X POST http://localhost:8080/api/laundry/silver/v1/powerCordInSocket \
  -F "photo=@cmd/api/powerCordInSocket/power2.png"

feedback: Geeft correct fail aan. 


# /api/laundry/silver/v1/rinseCycleMachineIsOn

✅
curl -X POST http://localhost:8080/api/laundry/silver/v1/rinseCycleMachineIsOn \
  -F "photo=@cmd/api/rinseCycleMachineIsOn/rinse1.png"

feedback: werkt

# /api/laundry/silver/v1/shippingBoltsRemoved
✅
curl -X POST http://localhost:8080/api/laundry/silver/v1/shippingBoltsRemoved \
  -F "photo=@cmd/api/shippingBoltsRemoved/bolts1.jpg"

feedback: failed, zoals hij moet want de schroeven zitten er nog in. 

✅
curl -X POST http://localhost:8080/api/laundry/silver/v1/shippingBoltsRemoved \
  -F "photo=@cmd/api/shippingBoltsRemoved/bolts2.png"

feedback: PASS, NICE!

# /api/laundry/silver/v1/levelIndicatorPresent

curl -X POST http://localhost:8080/api/laundry/silver/v1/levelIndicatorPresent \
  -F "photo=@cmd/api/levelIndicatorPresent/level1.png"

feedback: Werkt prima