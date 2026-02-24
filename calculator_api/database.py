"""Database engine, session management, and initialization helpers."""

from collections.abc import Generator

from sqlmodel import Session, SQLModel, create_engine

from .config import get_settings

_settings = get_settings()
engine = create_engine(_settings.database_url, echo=False)


def init_db() -> None:
    """Create all database tables defined in SQLModel metadata."""
    SQLModel.metadata.create_all(engine)


def get_db() -> Generator[Session, None, None]:
    """Yield a database session for use as a FastAPI dependency."""
    with Session(engine) as session:
        yield session
