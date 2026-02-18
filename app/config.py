"""Centralized application configuration via pydantic-settings."""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings loaded from environment variables with CALC_API_ prefix."""

    database_url: str = "sqlite:///calculator.db"
    jwt_secret_key: str = "dev-secret-key-change-me-in-production"
    jwt_algorithm: str = "HS256"
    access_token_expire_minutes: int = 30

    model_config = {"env_prefix": "CALC_API_", "env_file": ".env"}


settings = Settings()
