
import pafy

youtube = pafy.new("https://www.youtube.com/watch?v=RWhEUR0I9fo")

print(youtube.title)
print(youtube.getbestaudio().url)

