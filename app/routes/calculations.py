"""CRUD endpoints for calculation history."""

from fastapi import APIRouter, Depends, HTTPException
from sqlmodel import Session, col, select

from ..auth import CurrentUser
from ..core import add, divide, multiply, subtract
from ..database import get_session
from ..models import Calculation, CalculationCreate, CalculationResponse

router = APIRouter(prefix="/calculations", tags=["calculations"])

OPERATIONS: dict[str, callable] = {  # type: ignore[type-arg]
    "add": add,
    "sub": subtract,
    "mul": multiply,
    "div": divide,
}


@router.post("", response_model=CalculationResponse, status_code=201)
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


@router.get("", response_model=list[CalculationResponse])
def list_calculations(
    _user: CurrentUser, session: Session = Depends(get_session)
) -> list[Calculation]:
    """List all stored calculations."""
    stmt = select(Calculation).order_by(col(Calculation.created_at).desc())
    return list(session.exec(stmt).all())


@router.get("/{calculation_id}", response_model=CalculationResponse)
def get_calculation(
    calculation_id: int, _user: CurrentUser, session: Session = Depends(get_session)
) -> Calculation:
    """Get a specific calculation by ID."""
    calculation = session.get(Calculation, calculation_id)
    if not calculation:
        raise HTTPException(status_code=404, detail="Calculation not found")
    return calculation


@router.delete("/{calculation_id}", status_code=204)
def delete_calculation(
    calculation_id: int, _user: CurrentUser, session: Session = Depends(get_session)
) -> None:
    """Delete a calculation by ID."""
    calculation = session.get(Calculation, calculation_id)
    if not calculation:
        raise HTTPException(status_code=404, detail="Calculation not found")
    session.delete(calculation)
    session.commit()
