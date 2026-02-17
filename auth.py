"""JWT authentication module for the Calculator API."""

import os
from datetime import datetime, timedelta, timezone
from typing import Annotated

import bcrypt
from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jose import JWTError, jwt
from pydantic import BaseModel
from sqlmodel import Field, Session, SQLModel, select

# --- Configuration ---

SECRET_KEY = os.environ.get("JWT_SECRET_KEY", "dev-secret-key-change-me-in-production")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 30

# --- OAuth2 scheme ---

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/auth/token")

# --- Models ---


class User(SQLModel, table=True):
    """Database model for users."""

    id: int | None = Field(default=None, primary_key=True)
    username: str = Field(unique=True, index=True)
    hashed_password: str
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class UserCreate(BaseModel):
    """Request body for user registration."""

    username: str
    password: str


class UserResponse(BaseModel):
    """Response model for user info."""

    id: int
    username: str
    created_at: datetime


class Token(BaseModel):
    """Response model for an access token."""

    access_token: str
    token_type: str


class TokenData(BaseModel):
    """Decoded JWT token payload."""

    username: str | None = None


# --- Helpers ---


def verify_password(plain_password: str, hashed_password: str) -> bool:
    """Verify a plain password against its hash."""
    return bcrypt.checkpw(
        plain_password.encode("utf-8"),
        hashed_password.encode("utf-8"),
    )


def hash_password(password: str) -> str:
    """Hash a plain-text password."""
    return bcrypt.hashpw(
        password.encode("utf-8"),
        bcrypt.gensalt(),
    ).decode("utf-8")


def create_access_token(
    data: dict[str, str | datetime],
    expires_delta: timedelta | None = None,
) -> str:
    """Create a signed JWT access token."""
    to_encode: dict[str, str | datetime] = data.copy()
    expire = datetime.now(timezone.utc) + (expires_delta or timedelta(minutes=15))
    to_encode["exp"] = expire
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)  # type: ignore[no-any-return]


def authenticate_user(session: Session, username: str, password: str) -> User | None:
    """Validate credentials and return the user, or None if invalid."""
    statement = select(User).where(User.username == username)
    user = session.exec(statement).first()
    if not user or not verify_password(password, user.hashed_password):
        return None
    return user


# --- FastAPI dependency ---


def get_current_user(
    token: Annotated[str, Depends(oauth2_scheme)],
    session: Session,
) -> User:
    """Decode JWT and return the current authenticated user.

    This is meant to be used as a sub-dependency; the session must be
    injected by the caller via ``Depends``.
    """
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Could not validate credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    try:
        payload: dict[str, str] = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str | None = payload.get("sub")
        if username is None:
            raise credentials_exception
    except JWTError:
        raise credentials_exception

    user = session.exec(select(User).where(User.username == username)).first()
    if user is None:
        raise credentials_exception
    return user
