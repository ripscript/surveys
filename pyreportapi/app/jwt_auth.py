import jwt
from app.config import JWT_SECRET_KEY


def validate_token(context, is_secure: bool):
    """
    Return (success, message, code, claims, token)
    """
    metadata = dict(context.invocation_metadata())
    auth_header = metadata.get("authorization", "")

    if not auth_header:
        if is_secure:
            return False, "Header token tidak ditemukan", 401, {}, ""
        return True, "Tervalidasi", 200, {}, ""

    parts = auth_header.split(" ")
    if len(parts) != 2 or parts[0] != "Bearer":
        if is_secure:
            return False, "Format header token tidak valid", 401, {}, ""
        return True, "Tervalidasi", 200, {}, ""

    token = parts[1]

    try:
        claims = jwt.decode(token, JWT_SECRET_KEY, algorithms=["HS256"])
        return True, "Tervalidasi", 200, claims, token
    except jwt.ExpiredSignatureError:
        if is_secure:
            return False, "Token Kadaluarsa", 401, {}, ""
        return True, "Tervalidasi", 200, {}, ""
    except jwt.InvalidTokenError:
        if is_secure:
            return False, "Token Tidak Valid", 401, {}, ""
        return True, "Tervalidasi", 200, {}, ""