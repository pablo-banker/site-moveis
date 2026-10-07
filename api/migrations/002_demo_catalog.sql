-- Initial demonstration catalog; not inventory or sales data.
INSERT INTO categories(id,name,position) VALUES ('sofas','Sofás',0),
('poltronas','Poltronas',1),
('mesas','Mesas',2),
('cadeiras','Cadeiras',3),
('estantes-e-aparadores','Estantes e aparadores',4),
('camas','Camas',5);
INSERT INTO rooms(id,name,position) VALUES ('sala-de-estar','Sala de estar',0),
('sala-de-jantar','Sala de jantar',1),
('quarto','Quarto',2),
('escritorio','Escritório',3);
INSERT INTO products(id,name,category_id,room_id,price_cents,image,dimensions,material,label,finishes,position) VALUES 
('sofa-arco','Sofá Arco','sofas','sala-de-estar',429000,'https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=1000&q=85','220 × 92 × 78 cm','Linho e madeira maciça','Mais vendido',ARRAY['Linho natural','Cinza claro','Grafite'],0),
('poltrona-linha','Poltrona Linha','poltronas','sala-de-estar',189000,'https://images.unsplash.com/photo-1567538096630-e0c55bd6374c?auto=format&fit=crop&w=1000&q=85','76 × 82 × 80 cm','Tecido bouclé e carvalho','Novo',ARRAY['Bouclé claro','Cinza claro','Grafite'],1),
('mesa-encontro','Mesa Encontro','mesas','sala-de-jantar',269000,'https://images.unsplash.com/photo-1533090161767-e6ffed986c88?auto=format&fit=crop&w=1000&q=85','160 × 90 × 75 cm','Carvalho natural','',ARRAY['Carvalho natural','Nogueira'],2),
('cadeira-elo','Cadeira Elo','cadeiras','sala-de-jantar',79000,'https://images.unsplash.com/photo-1598300042247-d088f8ab3a91?auto=format&fit=crop&w=1000&q=85','48 × 52 × 80 cm','Madeira e tecido','',ARRAY['Carvalho natural','Nogueira'],3),
('aparador-plano','Aparador Plano','estantes-e-aparadores','sala-de-estar',219000,'https://images.unsplash.com/photo-1538688423619-a81d3f23454b?auto=format&fit=crop&w=1000&q=85','140 × 40 × 75 cm','Madeira natural','',ARRAY['Carvalho natural','Nogueira'],4),
('mesa-origem','Mesa de centro Origem','mesas','sala-de-estar',119000,'https://images.unsplash.com/photo-1499933374294-4584851497cc?auto=format&fit=crop&w=1000&q=85','90 × 60 × 35 cm','Carvalho natural','Novo',ARRAY['Carvalho natural','Nogueira'],5),
('cama-refugio','Cama Refúgio','camas','quarto',329000,'https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1000&q=85','158 × 198 × 100 cm','Madeira e tecido','',ARRAY['Linho natural','Grafite'],6),
('escrivaninha-traco','Escrivaninha Traço','mesas','escritorio',149000,'https://images.unsplash.com/photo-1518455027359-f3f8164ba6bd?auto=format&fit=crop&w=1000&q=85','120 × 60 × 75 cm','Madeira e aço','',ARRAY['Carvalho natural','Nogueira'],7);
