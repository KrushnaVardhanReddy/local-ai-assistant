from fastapi import APIRouter, Request, HTTPException
import json
import stripe
from auth import process_referral_reward
from config import config

router = APIRouter()

@router.post("/api/stripe_webhook")
async def handle_stripe_webhook(request: Request):
    payload = await request.body()
    sig_header = request.headers.get("stripe-signature")

    try:
        event = stripe.Webhook.construct_event(
            payload, sig_header, config.STRIPE_WEBHOOK_SECRET
        )
    except ValueError as e:
        # Invalid payload
        raise HTTPException(status_code=400, detail="Invalid payload")
    except stripe.error.SignatureVerificationError as e:
        # Invalid signature
        raise HTTPException(status_code=400, detail="Invalid signature")

    # The Stripe python SDK construct_event returns an Event object which behaves like a dictionary
    if hasattr(event, "type"):
        event_type = event.type
        data_object = event.data.object
    else:
        event_type = event.get("type")
        data_object = event.get("data", {}).get("object", {})

    # Check for subscription creation
    if event_type == "customer.subscription.created":
        if hasattr(data_object, "customer"):
            customer_id = data_object.customer
        else:
            customer_id = data_object.get("customer")

        if customer_id:
            try:
                await process_referral_reward(customer_id)
            except Exception as e:
                print(f"Error processing referral reward: {e}")

    return {"status": "success"}
