"""Calculator arithmetic endpoints: add, subtract, multiply, divide."""

from typing import Annotated

from fastapi import APIRouter, Depends, HTTPException

from ..auth import get_current_user
from ..core import add, divide, multiply, subtract
from ..models import CalculationRequest, ResultResponse, User

router = APIRouter(tags=["calculator"])

CurrentUser = Annotated[User, Depends(get_current_user)]


@router.post("/add", response_model=ResultResponse)
def api_add(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Add two numbers."""
    return ResultResponse(result=add(req.a, req.b))


@router.post("/subtract", response_model=ResultResponse)
def api_subtract(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Subtract b from a."""
    return ResultResponse(result=subtract(req.a, req.b))


@router.post("/multiply", response_model=ResultResponse)
def api_multiply(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Multiply two numbers."""
    return ResultResponse(result=multiply(req.a, req.b))


@router.post("/divide", response_model=ResultResponse)
def api_divide(req: CalculationRequest, _user: CurrentUser) -> ResultResponse:
    """Divide a by b."""
    try:
        return ResultResponse(result=divide(req.a, req.b))
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e)) from e
