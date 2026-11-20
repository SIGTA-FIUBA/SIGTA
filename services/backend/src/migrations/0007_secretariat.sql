-- The Secretariat and its platform permissions, which the backend checks by
-- their IDs to edit and publish procedure types (ADR-020, ADR-027). They come
-- with the code, so they are added here and not by an administrator. Areas,
-- roles and memberships are configuration of the organization, which the
-- administrators load in SIGTA.
-- Until this migration the server added them on every start, so a database
-- may have them already: the area keeps its name, and a permission with one
-- of these IDs becomes the platform one again if it was not, or was deleted.
INSERT INTO areas (id, number, name, system)
SELECT 'secretariat', coalesce(max(number), 0) + 1, 'Secretaría de Gestión Académica', true FROM areas
ON CONFLICT (id) DO UPDATE SET system = true, deleted_at = NULL, deleted_by = NULL;

INSERT INTO area_permissions (area_id, id, number, name, system)
SELECT 'secretariat', 'edit_procedure_types', coalesce(max(number), 0) + 1, 'Editar tipos de trámite', true
FROM area_permissions WHERE area_id = 'secretariat'
ON CONFLICT (area_id, id) DO UPDATE SET name = excluded.name, system = true, deleted_at = NULL, deleted_by = NULL
WHERE NOT area_permissions.system OR area_permissions.deleted_at IS NOT NULL;

INSERT INTO area_permissions (area_id, id, number, name, system)
SELECT 'secretariat', 'publish_procedure_types', coalesce(max(number), 0) + 1, 'Publicar tipos de trámite', true
FROM area_permissions WHERE area_id = 'secretariat'
ON CONFLICT (area_id, id) DO UPDATE SET name = excluded.name, system = true, deleted_at = NULL, deleted_by = NULL
WHERE NOT area_permissions.system OR area_permissions.deleted_at IS NOT NULL;
