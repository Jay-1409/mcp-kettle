from fastapi import APIRouter

router = APIRouter()


@router.get("/users")
def list_users(limit: int = 20):
    return {"limit": limit, "users": []}


@router.get("/users/{user_id}")
def get_user(user_id: int):
    return {"user_id": user_id}


@router.post("/users/{user_id}/activate")
def activate_user(user_id: int):
    return {"user_id": user_id, "active": True}


@router.delete("/users/{user_id}")
def delete_user(user_id: int):
    return {"deleted": user_id}


@router.get("/users/{user_id}/sessions")
def list_sessions(user_id: int, active: bool = True):
    return {"user_id": user_id, "active": active, "sessions": []}
