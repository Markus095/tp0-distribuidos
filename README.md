### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°6 explicación de Solución:
El protocolo ya estaba planteado con la idea de manejar varias apuestas en un solo mensaje, así que no fue necesario modificarlo.
Se modificó al cliente para poder levantar los archivos del csv en el docker volume y enviar las apuestas en varios mensajes al servidor. Cada mensaje puede tener hasta "batch: maxAmount" apuestas, pero (generalmente el último) pueden tener menos y eso no genera ningún problema. También se lo separó en varios archivos para facilitar la lectura y poder mantener la separación de responsabilidades.
El servidor se modificó para esperar varios mensajes de un mismo cliente y para poder procesar y guardar las varias apuestas

### Dificultades
-Inicialmente moví los directorios a una carpeta "/datasets" y funcionaba bien, pero como las pruebas no lo encontraban tuve que modificar toda la lógica de la creacion de los volumenes y en lugar de pasarle un archivo ahora los clientes reciben el directorio.