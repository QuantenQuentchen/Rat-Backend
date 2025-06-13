import jwt
import time
import uuid

# Load your private key
with open("private.pem", "rb") as f:
    private_key = f.read()

def generate_jwt():
    payload = {
        "sub": "test",
        "aud": "governance-backend",
        "iss": "gov-bot",
        "iat": int(time.time()),
        "exp": int(time.time()) + 300,  # token valid for 5 minutes
        "jti": str(uuid.uuid4())        # unique token ID
    }
    token = jwt.encode(payload, private_key, algorithm="RS256")
    return token

if __name__ == "__main__":
    token = generate_jwt()
    print("Generated JWT:")
    print(token)
