# language: es
Característica: Autenticación con proveedor abstraído (US06)
  Como persona usuaria,
  quiero iniciar sesión,
  para acceder solo a lo que me corresponde.

  Antecedentes:
    Dado que solo se admiten cuentas del dominio "fi.uba.ar"
    Y que SIGTA tiene registrada a "titular@fi.uba.ar" con perfil titular
    Y que SIGTA tiene registrada a "operativo@fi.uba.ar" con perfil Secretaría operativo

  Escenario: El perfil sale de SIGTA, no del proveedor
    Cuando "operativo@fi.uba.ar" ingresa a través del proveedor de identidad
    Entonces se abre una sesión
    Y su perfil es Secretaría operativo

  Escenario: Una identidad que SIGTA no conoce no obtiene sesión
    Cuando "nadie@fi.uba.ar" ingresa a través del proveedor de identidad
    Entonces no se abre ninguna sesión

  Esquema del escenario: Solo ingresan cuentas institucionales
    Cuando "<correo>" ingresa a través del proveedor de identidad
    Entonces no se abre ninguna sesión
    Y se le pide que ingrese con su cuenta institucional

    Ejemplos:
      | correo                       |
      | titular@example.com          |
      | operativo@evilfi.uba.ar      |
      | operativo@fi.uba.ar.evil.com |

  Escenario: Un ingreso que no inició el navegador se rechaza
    Cuando llega una respuesta del proveedor con un estado distinto al del ingreso iniciado
    Entonces no se abre ninguna sesión

  Escenario: La sesión vence por inactividad
    Dado que "titular@fi.uba.ar" tiene una sesión abierta
    Cuando pasan 30 minutos sin pedidos
    Entonces la sesión vence
    Y tiene que volver a iniciar sesión

  Escenario: La actividad mantiene la sesión abierta
    Dado que "titular@fi.uba.ar" tiene una sesión abierta
    Cuando hace un pedido cada 29 minutos
    Entonces la sesión sigue abierta

  Escenario: Un titular ingresa con un enlace por correo
    Cuando "titular@fi.uba.ar" pide un enlace de ingreso
    Entonces recibe por correo un enlace de un solo uso
    Y al abrirlo se abre una sesión

  Escenario: El enlace no se puede reutilizar
    Dado que "titular@fi.uba.ar" ya usó su enlace de ingreso
    Cuando lo vuelve a abrir
    Entonces no se abre ninguna sesión

  Escenario: El enlace vence
    Dado que "titular@fi.uba.ar" pidió un enlace de ingreso
    Cuando pasan 15 minutos
    Entonces el enlace ya no abre una sesión

  Escenario: El enlace solo se envía a correos institucionales
    Cuando "titular@example.com" pide un enlace de ingreso
    Entonces se le pide que use su correo institucional
    Y no se envía ningún correo

  Escenario: El pedido de enlace no revela quién tiene cuenta
    Cuando "nadie@fi.uba.ar" pide un enlace de ingreso
    Entonces la respuesta es la misma que para una cuenta existente
    Y no se envía ningún correo

  Escenario: El personal ingresa solo a través del proveedor de identidad
    Cuando "operativo@fi.uba.ar" pide un enlace de ingreso
    Entonces no se envía ningún correo
