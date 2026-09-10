from pydantic import model_validator


@model_validator(mode="after")
def _normalize_email(self):
    if isinstance(self.email, str):
        self.email = self.email.strip().lower()
    return self
