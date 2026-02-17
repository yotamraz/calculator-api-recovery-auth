#!/usr/bin/env python3
"""A very simple calculator REST API server."""

from collections.abc import Generator
from contextlib import asynccontextmanager
from datetime import datetime, timedelta, timezone
from typing import Annotated, AsyncGenerator

import uvicorn
from fastapi import Depends, FastAPI, HTTPException, status
from fastapi.security import OAuth2PasswordRequestForm
from pydantic import BaseModel
from sqlmodel import Field, Session, SQLModel, col, create_engine, select

from auth import (
    ACCESS_TOKEN_EXPIRE_MINUTES,
    Token,
    User,
    UserCreate,
    UserResponse,
    authenticate_user,
    create_access_token,
    get_current_user,
    hash_password,
    oauth2_scheme,
)
from core import add, divide, multiply, subtract

# --- Database Setup ---

DATABASE_URL = "sqlite:///calculator.db"
engine = create_engine(DATABASE_URL, echo=False)


def get_session() -> Generator[Session, None, None]:
    """Yield a database session."""
    with Session(engine) as session:
        yield session


# --- Models ---


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


class CalculationRequest(BaseModel):
    """Request body for calculation endpoints."""

    a: float
    b: float


class ResultResponse(BaseModel):
    """Response body for calculation endpoints."""

    result: float


# --- App Setup ---


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Create database tables on startup."""
    SQLModel.metadata.create_all(engine)
    yield


app = FastAPI(title="Calculator API", version="0.1.0", lifespan=lifespan)

OPERATIONS: dict[str, callable] = {  # type: ignore[type-arg]
    "add": add,
    "sub": subtract,
    "mul": multiply,
    "div": divide,
}


# --- Auth dependency ---


def _get_current_user(
    token: Annotated[str, Depends(oauth2_scheme)],
    session: Session = Depends(get_session),
) -> User:
    """Resolve the current authenticated user from the JWT + DB session."""
    return get_current_user(token, session)


CurrentUser = Annotated[User, Depends(_get_current_user)]


# --- Health Check ---


class HealthResponse(BaseModel):
    """Response model for the health check endpoint."""

    status: str
    version: str


@app.get("/health", response_model=HealthResponse)
def health_check() -> HealthResponse:
    """Return the health status of the service."""
    return HealthResponse(status="ok", version=app.version)


# --- Auth Endpoints ---


@app.post("/auth/register", response_model=UserResponse, status_code=201)
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


@app.post("/auth/token", response_model=Token)
def login(
    form_data: Annotated[OAuth2PasswordRequestForm, Depends()],
    session: Session = Depends(get_session),
) -> Token:
    """Authenticate and return a JWT access token.

    Uses the standard OAuth2 password flow (form fields: ``username``, ``password``).
    """
    user = authenticate_user(session, form_data.username, form_data.password)
    if not user:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Incorrect username or password",
            headers={"WWW-Authenticate": "Bearer"},
        )
    access_token = create_access_token(
        data={"sub": user.username},
        expires_delta=timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES),
    )
    return Token(access_token=access_token, token_type="bearer")


# --- Calculator Endpoints ---


@app.post("/add", response_model=ResultResponse)
def api_add(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Add two numbers."""
    return ResultResponse(result=add(req.a, req.b))


@app.post("/subtract", response_model=ResultResponse)
def api_subtract(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Subtract b from a."""
    return ResultResponse(result=subtract(req.a, req.b))


@app.post("/multiply", response_model=ResultResponse)
def api_multiply(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Multiply two numbers."""
    return ResultResponse(result=multiply(req.a, req.b))


@app.post("/divide", response_model=ResultResponse)
def api_divide(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Divide a by b."""
    try:
        return ResultResponse(result=divide(req.a, req.b))
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e)) from e


# --- CRUD Endpoints for Calculations ---


@app.post("/calculations", response_model=CalculationResponse, status_code=201)
def create_calculation(
    req: CalculationCreate, _user: CurrentUser, session: Session = Depends(get_session)
) -> Calculation:
    """Create and store a new calculation."""
    if req.operation not in OPERATIONS:
        raise HTTPException(
            status_code=400, detail=f"Unknown operation: {req.operation}. Use: {list(OPERATIONS)}"
        )

    try:
        result = OPERATIONS[req.operation](req.a, req.b)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e)) from e

    calculation = Calculation(operation=req.operation, a=req.a, b=req.b, result=result)
    session.add(calculation)
    session.commit()
    session.refresh(calculation)
    return calculation


@app.get("/calculations", response_model=list[CalculationResponse])
def list_calculations(
    _user: CurrentUser, session: Session = Depends(get_session)
) -> list[Calculation]:
    """List all stored calculations."""
    stmt = select(Calculation).order_by(col(Calculation.created_at).desc())
    return list(session.exec(stmt).all())


@app.get("/calculations/{calculation_id}", response_model=CalculationResponse)
def get_calculation(
    calculation_id: int, _user: CurrentUser, session: Session = Depends(get_session)
) -> Calculation:
    """Get a specific calculation by ID."""
    calculation = session.get(Calculation, calculation_id)
    if not calculation:
        raise HTTPException(status_code=404, detail="Calculation not found")
    return calculation


@app.delete("/calculations/{calculation_id}", status_code=204)
def delete_calculation(
    calculation_id: int, _user: CurrentUser, session: Session = Depends(get_session)
) -> None:
    """Delete a calculation by ID."""
    calculation = session.get(Calculation, calculation_id)
    if not calculation:
        raise HTTPException(status_code=404, detail="Calculation not found")
    session.delete(calculation)
    session.commit()


def main() -> None:
    """Run the server."""
    uvicorn.run(app, host="0.0.0.0", port=8000)


if __name__ == "__main__":
    main()
