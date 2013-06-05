import sys, os, urllib, urllib2, json, cookielib, re

if len(sys.argv) < 5:
	print("Usage: python export.py host username password export_dir")
	sys.exit(1)

host = sys.argv[1]
username = sys.argv[2]
password = sys.argv[3]
export_dir = sys.argv[4]

login_url = "http://%s/api.php?action=login&lgname=%s&lgpassword=%s&format=json" % (host, username, password)
list_url = "http://%s/api.php?format=json&action=query&list=allpages&aplimit=max" % host

opener = urllib2.build_opener(urllib2.HTTPCookieProcessor(cookielib.CookieJar()))

# Login and confirm token
result = json.loads(opener.open(login_url, " ").readline());
opener.open(login_url, "lgtoken=%s" % result.get('login').get('token'))

# Make sure directory exists
if not os.path.isdir(export_dir):
	os.makedirs(export_dir)

# Get article list
result = json.loads(opener.open(list_url).readline())
for page in result.get('query').get('allpages'):
	title = page.get('title')
	params = {"action":"raw","title":title}
	article_url = "http://%s/index.php?%s" % (host, urllib.urlencode(params))
	scrubbed_title = re.sub(r'[^A-Za-z0-9_]', r'_', title)
	output_file = os.path.join(export_dir, "%s.txt" % scrubbed_title)
	print "%s: %s to %s" % (title, article_url, output_file)
	file = open(output_file, 'wb')
	try:
		for line in opener.open(article_url):
			file.write(line)
		file.write('\n')
	finally:
		file.close()
