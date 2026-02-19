"""Auth endpoints: user registration and token issuance."""

from datetime import datetime, timedelta, timezone
from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import OAuth2PasswordRequestForm
from sqlmodel import Session, select

from ..auth import authenticate_user, create_access_token, hash_password
from ..config import get_settings
from ..database import get_session
from ..models import Token, User, UserCreate, UserResponse

router = APIRouter()


@router.post("/register", response_model=UserResponse, status_code=201)
def register(req: UserCreate, session: Session = Depends(get_session)) -> User:
    """Register a new user."""
    existing = session.exec(select(User).where(User.username == req.username)).first()
    if existing:
        raise HTTPException(status_code=400, detail="Username already taken")

    user = User(
        username=req.username,
        hashed_password=hash_password(req.password),
        created_at=datetime.now(timezone.utc),
    )
    session.add(user)
    session.commit()
    session.refresh(user)
    return user


@router.post("/token", response_model=Token)
def login(
    form_data: Annotated[OAuth2PasswordRequestForm, Depends()],
    session: Session = Depends(get_session),
) -> Token:
    """Authenticate and return a JWT access token."""
    user = authenticate_user(session, form_data.username, form_data.password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Bearer"},
        )
    settings = get_settings()
    access_token = create_access_token(
        data={"sub": user.username},
        expires_delta=timedelta(minutes=settings.access_token_expire_minutes),
    )
    return Token(access_token=access_token, token_type="bearer")
