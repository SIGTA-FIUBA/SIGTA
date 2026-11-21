// Texts that say what SIGTA is and how it is used. The landing page and the
// home page share them, so both say the same thing.

export type ContentItem = { title: string; text: string }

// What each kind of person does in SIGTA, in the order they do it.
export const applicantHowTo: ContentItem[] = [
  {
    title: 'Consulta del catálogo',
    text: 'Cada trámite detalla sus requisitos, plazos, condiciones de admisibilidad y la documentación necesaria.',
  },
  {
    title: 'Presentación de la solicitud',
    text: 'La solicitud se completa en línea con los datos y documentos requeridos. La plataforma verifica que no falte información antes del envío.',
  },
  {
    title: 'Seguimiento',
    text: 'Cada trámite indica la instancia en la que se encuentra y lo que se espera a continuación. Las novedades se notifican por correo.',
  },
]

export const principles: ContentItem[] = [
  {
    title: 'Las decisiones son de una persona',
    text: 'Las herramientas de asistencia sugieren y señalan. La resolución de cada trámite corresponde al área competente.',
  },
  {
    title: 'Registro completo',
    text: 'Cada acción queda registrada con su autor y su fecha, y puede consultarse en el historial del trámite.',
  },
  {
    title: 'Registro de lo actuado en otros sistemas',
    text: 'El personal confirma en SIGTA lo que hizo en el SIU Guaraní y en el expediente electrónico. SIGTA no se conecta con ellos ni los reemplaza.',
  },
]

// The core of SIGTA, whatever procedures it holds.
export const pillars: ContentItem[] = [
  {
    title: 'Un solo lugar para cada trámite',
    text: 'Cada trámite se inicia en línea, con sus requisitos a la vista y la documentación en formato digital, sin correos ni planillas.',
  },
  {
    title: 'Seguimiento de cada instancia',
    text: 'El titular y la Secretaría ven en qué instancia está el trámite, quién lo tiene a cargo y qué se espera a continuación.',
  },
  {
    title: 'Trámites configurables',
    text: 'Los trámites se definen por configuración: incorporar uno nuevo no requiere desarrollo, y cada cambio queda versionado.',
  },
]
