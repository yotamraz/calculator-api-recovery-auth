"""SQLModel table models and Pydantic request/response schemas."""

from datetime import datetime, timezone

from pydantic import BaseModel
from sqlmodel import Field, SQLModel


# --- SQLModel Table Models ---


class User(SQLModel, table=True):
    """Database model for users."""

    id: int | None = Field(default=None, primary_key=True)
    username: str = Field(unique=True, index=True)
    hashed_password: str
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class Calculation(SQLModel, table=True):
    """Database model for stored calculations."""

    id: int | None = Field(default=None, primary_key=True)
    operation: str
    a: float
    b: float
    result: float
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


# --- Pydantic Schemas ---


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


class CalculationBase(SQLModel):
    """Base model for calculations."""

    operation: str
    a: float
    b: float
    result: float


class CalculationCreate(BaseModel):
    """Request body for creating a calculation."""

    operation: str
    a: float
    b: float


class CalculationResponse(CalculationBase):
    """Response model for a calculation."""

    id: int
    created_at: datetime


class CalculationRequest(BaseModel):
    """Request body for calculation endpoints."""

    a: float
    b: float


class ResultResponse(BaseModel):
    """Response body for calculation endpoints."""

    result: float


class HealthResponse(BaseModel):
    """Response model for the health check endpoint."""

    status: str
    version: str
