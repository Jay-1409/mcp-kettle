from fastapi import FastAPI

from admin import router as admin_router
from billing import router as billing_router
from projects import router as projects_router
from users import router as users_router

app = FastAPI()
app.include_router(users_router)
app.include_router(projects_router)
app.include_router(billing_router)
app.include_router(admin_router)


@app.get("/health")
def health():
    return {"message": "healthy"}
