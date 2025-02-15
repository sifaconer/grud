# grud - Generador de CRUD en Golang 🚀

**grud** es una herramienta CLI para generar automáticamente APIs REST en **Golang** usando **Bun ORM** y **PostgreSQL**.  
Permite crear, extender y modificar CRUDs dinámicamente sin afectar código existente, utilizando **AST** y **templates**.

## ✨ Características
- 🔧 **Generación Automática de CRUDs** a partir de tablas de la base de datos.
- 🛠 **Extensión de Código Existente** sin sobrescribir archivos manualmente.
- 🔍 **Filtros Avanzados**: Soporta `WHERE`, `OR`, `BETWEEN`, `GROUP BY`, `JOINs` y más.
- 🔑 **Soporte para RBAC** con Casbin.
- 📄 **Generación Automática de Swagger/OpenAPI**.

## 📌 Instalación
```sh
go install github.com/sifaconer/grud@latest
