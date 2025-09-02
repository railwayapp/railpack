from pdf2image import convert_from_path, convert_from_bytes
from pydub import AudioSegment
from pydub.playback import play
import cairo

# pdf2image
images = convert_from_path('./test.pdf')
print(images)

# pydub
sound = AudioSegment.from_file("test.mp3", format="mp3")
print(sound)

# pycairo
surface = cairo.SVGSurface("example.svg", 200, 200)
print(surface)
print("PyCairo version:", cairo.version)

print("Hello from Python with system deps")
