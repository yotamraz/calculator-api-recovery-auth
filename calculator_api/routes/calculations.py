"""Calculator operation endpoints and CRUD history routes."""

from fastapi import APIRouter, HTTPException

from ..core import add, divide, multiply, subtract
from ..models import CalculationRequest, ResultResponse

router = APIRouter()


# --- Calculator Operation Endpoints ---


@router.post("/add", response_model=ResultResponse)
def api_add(req: CalculationRequest) -> ResultResponse:
    """Add two numbers."""
    return ResultResponse(result=add(req.a, req.b))


@router.post("/subtract", response_model=ResultResponse)
def api_subtract(req: CalculationRequest) -> ResultResponse:
    """Subtract b from a."""
    return ResultResponse(result=subtract(req.a, req.b))


@router.post("/multiply", response_model=ResultResponse)
def api_multiply(req: CalculationRequest) -> ResultResponse:
    """Multiply two numbers."""
    return ResultResponse(result=multiply(req.a, req.b))


@router.post("/divide", response_model=ResultResponse)
def api_divide(req: CalculationRequest) -> ResultResponse:
    """Divide a by b."""
    try:
        return ResultResponse(result=divide(req.a, req.b))
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e)) from e
