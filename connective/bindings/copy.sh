#::run : subdirs (store server manager devicemanager pyconsole btmanager)
#::end

#::gen repcsv 'rm ../../$1/bindings.py' (subdirs)
rm ../../server/bindings.py
rm ../../manager/bindings.py
rm ../../devicemanager/bindings.py
rm ../../pyconsole/bindings.py
rm ../../btmanager/bindings.py
#::end

#::gen repcsv 'cp ./bindings.py ../../$1/bindings.py' (subdirs)
cp ./bindings.py ../../server/bindings.py
cp ./bindings.py ../../manager/bindings.py
cp ./bindings.py ../../devicemanager/bindings.py
cp ./bindings.py ../../pyconsole/bindings.py
cp ./bindings.py ../../btmanager/bindings.py
#::end

