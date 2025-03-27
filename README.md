### Ejercicio N°1 explicación de solución:
Se creó un script muy sencillo de Bash, que solo se encarga de imprimir los parámetros por consola y luego enviarselos a un archivo de python. Este archivo llamado "mi-generador.py" es el encargado de:
1-Validar que son parámetros válidos para generar un nuevo archivo de compose.
2-Abrir un archivo nuevo, con el nombre especificado.
3-Agregar al servidor al archivo, siguiendo el formato del servidor de ejemplo.
4-Agregar tantos clientes como se le especificaron, siguiendo el formato anterior y asegurándose de cambiar sus ids y nombres.
5-Agregar la red a la cual están todos conectados.
6-Si pudo agregar todo y guardar el archivo con éxito, notifica al usuario y el programa termina.

### Dificultades Encontradas
-Hubo ciertas complicaciones a la hora de parsear el archivo, a veces las líneas quedaban desfasadas.


### Ejercicio N°2 explicación de solución:
Se modificó el archivo de compose (y por lo tanto el generador) para que el servidor y el/los cliente/s tengan un volumen por el cual se pueda acceder a sus respectivas configuraciones.
Además ya no se especifica su logging level desde el compose.

### Dificultades Encontradas
-Hubo cierta dificultad en encontrar el path correcto de los archivos de configuración a los volúmenes. Al principio intentaba con paths absolutos y no terminaba de entender por qué fallaba. Al usar paths relativos se solucionó.


### Ejercicio N°3 explicación de la solución:
El script funciona de la siguiente manera:
1- Crea un container temporal con netcat instalado.
2- Duerme un par de segundos para esperar la instalación.
3- Envía un mensaje al servidor a través de la red y luego captura su respuesta en una variable usando docker exec.
4- Elimina el container.
5- Compara la respuesta con el texto enviado. Si son iguales, imprime por consola que la prueba fue un éxito. Caso contrario, imprime que fue un fracaso.

### Dificultades encontradas
-Al principio intenté hacer la prueba a través de los clientes, pero terminé decidiendo que crear un container nuevo era más sencillo y práctico.
-Fue difícil comprender la lógica de conectarme a la red y obtener la ip del servidor. Tuve que leer documentación de docker network.
-Tardé un poco en entender que tenía que conectarme a través del puerto especificado en la configuración del servidor.
-Tuve que aprender a usar variables y ifs en bash, que nunca me había hecho falta hasta ahora.

### Ejercicio N°4 explicación de la solución:
Se le agregó al servidor un nuevo parámetro: running, el cual es revisado en lugar del while true original. Se inicializa con un valor positivo, pero cambia cuando recibe el sigterm. Cuando eso sucede, se corta el loop y trata de cerrar todos los sockets y luego finaliza. 
Por otro lado, se le agregó un waitgroup al cliente para revisar que las conexiones se cierran antes de terminar el programa y un campo "done" que se usa para avisarle a todas las rutinas cuando se solicita el cierre del cliente. También se le añadió un signal handler que funciona en una rutina separada para manejar la secuencia de cerrado y la secuencia de cierre, que cierra todas las conexiones.

### Dificultades encontradas
Al principio me costó encontrar el formato apropiado para el logging, así que tuve varios commits innecesarios para probar como funcionaban distintos formatos con los tests.
Durante bastante tiempo le pasaba una cantidad erronea de parametros al handler del SIGTERM del servidor, así que el cierre no funcionaba correctamente.


### Ejercicio N°5 explicación de solución:
Se creó un protocolo para la comunicación entre el cliente y el servidor.
Este consiste en el envío de mensajes encodeados a bytes a través del socket TCP. Los mensajes provenientes del cliente tienen la siguiente estructura:
## Header
Tiene un tamaño fijo de 6 bytes, contiene:
-Número de agencia (4 bytes) representado como un entero de 32 bits en formato big-endian.
-Número de apuestas (2 bytes) representado como un entero sin signo de 16 bits en formato big-endian.

## Body
Este contiene la información de las apuestas. Cada apuesta tiene un tamaño fijo de 146 bytes. 
# Campos de la apuesta:
-Nombre ( 64 bytes) se trata de un string que se rellena con espacios cuando es mas corto.
-Apellido ( 64 bytes) al igual que el nombre, un string que se rellena con espacios cuando es mas corto.
-Documento ( 8 bytes), rellenado con espacios cuando es más corto.
-Fecha de Nacimiento (8 bytes), en formato ISO, YYYYMMDD
-Número apostado (2 bytes), entero sin signo de 16 bits.

Por lo tanto el tamaño total de un mensaje es Header(6 bytes) + Tamaño de una apuesta

El Cliente parsea este mensaje con las variables de la apuesta que se le pasan por parámetro.
Cuando el Servidor lo recibe, almacena el ID del cliente y luego decodifica cada apuesta, usando el ID del header para el campo de la agencia. Finalmente almacena las apuestas y envía un mensaje de 2 bytes que dice ("OK) como ACK, al recibirlo el cliente termina.

### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia
En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8:

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega
Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).
