### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).

### Ejercicio N°2 explicación de solución:
Se modificó el archivo de compose (y por lo tanto el generador) para que el servidor y el/los cliente/s tengan un volumen por el cual se pueda acceder a sus respectivas configuraciones.
Además ya no se especifica su logging level desde el compose.

### Dificultades Encontradas
-Hubo cierta dificultad en encontrar el path correcto de los archivos de configuración a los volúmenes. Al principio intentaba con paths absolutos y no terminaba de entender por qué fallaba. Al usar paths relativos se solucionó.