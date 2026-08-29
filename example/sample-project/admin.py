from fastapi import APIRouter

router = APIRouter()


@router.get("/admin/users")
def admin_users(limit: int = 100):
    return {"limit": limit, "users": []}


@router.get("/admin/audit")
def audit_log(offset: int = 0):
    return {"offset": offset, "events": []}


@router.post("/admin/maintenance/{job_id}")
def run_maintenance(job_id: int, dry_run: bool = True):
    return {"job_id": job_id, "dry_run": dry_run}


@router.delete("/admin/sessions/{session_id}")
def revoke_session(session_id: int):
    return {"revoked": session_id}
