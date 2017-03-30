createmarkdownreadme:
	echo '# Routing\n' > readme.md
	sed 's/^\/\/\(\s\|$$\)//g' doc.go | head -n -1 >> readme.md

