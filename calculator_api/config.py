"""Centralized configuration using pydantic-settings."""

from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables and/or .env file."""

    database_url: str = "sqlite:///calculator.db"
    jwt_secret_key: str = "dev-secret-key-change-me-in-production"
    jwt_algorithm: str = "HS256"
    access_token_expire_minutes: int = 30

    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")


@lru_cache
def get_settings() -> Settings:
    """Return a cached Settings instance."""
    return Settings()
