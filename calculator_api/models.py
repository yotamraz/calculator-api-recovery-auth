"""SQLModel database models and Pydantic request/response schemas."""

from datetime import datetime

from pydantic import BaseModel
from sqlmodel import Field, SQLModel


# --- Calculation Models ---


class CalculationBase(SQLModel):
    """Base model for calculations."""

    operation: str
    a: float
    b: float
    result: float


class Calculation(CalculationBase, table=True):
    """Database model for stored calculations."""

    id: int | None = Field(default=None, primary_key=True)
    created_at: datetime = Field(default_factory=datetime.utcnow)


class CalculationCreate(BaseModel):
    """Request body for creating a calculation."""

    operation: str
    a: float
    b: float


class CalculationResponse(CalculationBase):
    """Response model for a calculation."""

    id: int
    created_at: datetime


# --- Calculator Request/Response Schemas ---


class CalculationRequest(BaseModel):
    """Request body for calculation endpoints."""

    a: float
    b: float


class ResultResponse(BaseModel):
    """Response body for calculation endpoints."""

    result: float


# --- Health Check Schema ---


class HealthResponse(BaseModel):
    """Response model for the health check endpoint."""

    status: str
    version: str
