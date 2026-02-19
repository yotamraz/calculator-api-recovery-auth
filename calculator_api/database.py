"""Database engine creation, session management, and initialization."""

from collections.abc import Generator

from sqlmodel import Session, SQLModel, create_engine

from .config import get_settings

engine = create_engine(get_settings().database_url, echo=False)


def init_db() -> None:
    """Create all database tables. Called at application startup."""
    SQLModel.metadata.create_all(engine)


def get_session() -> Generator[Session, None, None]:
    """Yield a database session for use as a FastAPI dependency."""
    with Session(engine) as session:
        yield session
