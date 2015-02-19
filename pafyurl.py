
import pafy

print(pafy.new("https://www.youtube.com/watch?v=RWhEUR0I9fo").getbestaudio().url)

