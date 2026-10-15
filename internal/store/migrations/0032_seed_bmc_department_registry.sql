-- 0032: the register of BMC departments and where each one's record lives.
--
-- Until now the platform could name the ward desk for three categories and
-- nothing else. Asked who repairs a footbridge, who replaces a street light or
-- who answers for a leaning compound wall, it had no row to consult — not
-- because BMC hides it, but because the portal's own index was never read.
--
-- It is read now. BMC publishes a Department List at
-- /irj/portal/anonymous/qlDeptManualsPage carrying every department in Marathi
-- and English, each with up to three links: Overview, RTI and Docs. These are
-- the 87 departments it lists, with the two links worth keeping.
--
-- This table is a register, not a routing rule. It records that a department
-- exists and where its Section 4 handbook is published. It does NOT say which
-- department answers for a given complaint: that claim needs evidence this
-- migration does not have, and the street-light case below shows why.
--
-- Three things learned reading it, recorded so the next person does not repeat
-- the work:
--
-- 1. The RTI link is reliable; the Docs link usually is not. `qlCeb` (RTI)
--    returns the eighteen Bridges handbook chapters. `qlBridges` (Docs)
--    returns only the portal's homepage boilerplate because its inner iView
--    never renders. Both are stored; prefer rti_url.
--
-- 2. The portal must be asked with a browser User-Agent. It answers anything
--    else with "Could not open iView. The iView is not compatible with your
--    browser" and HTTP 200, which reads as a dead page and is not one. Several
--    earlier attempts at this index were abandoned over exactly that.
--
-- 3. There is no URL pattern to guess. Ward handbooks sit under 'RTI Manuals',
--    Bridges under 'RTI manuals', and Mechanical & Electrical under a folder
--    spelling the name 'Mechanical and Electrical' rather than with an
--    ampersand. Each department's RTI page has to be read for its own paths.
--
-- One defect in the source, left as found: Chief Engineer (Bridges) and Chief
-- Engineer (Building Maintenance) are published with the same Overview link.
-- The Overview column is therefore not stored at all — it cannot be a key.

CREATE TABLE IF NOT EXISTS bmc_departments (
  id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name_en   text NOT NULL UNIQUE,
  name_mr   text,
  rti_url   text,
  docs_url  text,
  source_ref text NOT NULL
    DEFAULT 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlDeptManualsPage',
  retrieved_at timestamptz NOT NULL DEFAULT '2026-10-15T00:00:00Z',
  CONSTRAINT has_some_link CHECK (rti_url IS NOT NULL OR docs_url IS NOT NULL)
);

COMMENT ON TABLE bmc_departments IS
  'Every department BMC lists, with its published Section 4 handbook. A register of what exists, not a routing rule.';
COMMENT ON COLUMN bmc_departments.rti_url IS
  'Section 4(1)(b) handbook index. Preferred over docs_url, which often fails to render.';

INSERT INTO bmc_departments (name_en, name_mr, rti_url, docs_url)
VALUES
  ('ACWORTH Municipal Hospital for Leprosy', 'ऍकवर्थ महानगरपालिकाकुष्ठरुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlacworthmunhosp', NULL),
  ('Administrative Officer (School) P/South Ward', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlAOSPSW', NULL),
  ('Assessor & Collector', 'करनिर्धारणव संकलन', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlandc', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlandcdocs'),
  ('Assessment And Collection - Election', 'करनिर्धारण व संकलन - निवडणूक', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlandcele', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlelectiondocs'),
  ('Assessment & Collection - Octroi', 'करनिर्धारण व संकलन - जकात', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlacoctroi', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlaco'),
  ('Assistant Commissioner - Planning', 'सहाय्यक आयुक्त - नियोजन', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrtiacplanning', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlacpanonymous'),
  ('Assistant Commissioner - Removal of Encroachments', 'सहाय्यकआयुक्त - अतिक्रमण निर्मूलन', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlacre', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlacredocs'),
  ('B.Y.L.Nair Hospital & T.N.M.College', 'बी वाय नायर रूग्णालय व टि एन एम वैद्यकिय महाविद्यालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlbylnairhosp', NULL),
  ('Bhagwati Hospital', 'भगवती रुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBhagwatiHosp', NULL),
  ('Central Archives Cell', 'मध्‍यवर्ती पुराभिलेख कक्ष', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcac', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcarchivecelldoc'),
  ('Central Purchase Department', 'मध्यवर्ती खरेदी खाते', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcentpura', NULL),
  ('Chief Accountant - Finance(Treasury)', 'प्रमुख लेखापाल - वित्त (कोषागार)', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcafinance', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcafinancedocs'),
  ('Chief Accountant - W.S.S.D.', 'प्रमुख लेखापाल - पा.पु.म.नि.', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrticawssd', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlwssd'),
  ('Chief Engineer (Building Maintenance) Department', 'प्रमुख अभियंता (इमारत परिरक्षण)', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcityeng11', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlChiefBuildM'),
  ('Chief Engineer (Bridges)', 'प्रमुख अभियंता पूल', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlCeb', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlBridges'),
  ('Chief Engineer (Coastal Road Project)', 'प्रमुख अभियंता (किनारी रस्ता प्रकल्प)', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlChiefEnggCoaRP', NULL),
  ('Chief Engineer - Development Plan', 'प्रमुख अभियंता - विकास नियोजन', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcedp', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcedpdocs'),
  ('Chief Engineer - Mechanical and Electrical', 'प्रमुख अभिंयता - यांत्रिकी आणि विद्युत', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrticemne', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmechdocs'),
  ('प्रमुख अभियंता - मलनि:सारणप्रकल्प Chief Engineer - Sewerage Project', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcesewerageproj', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcesp'),
  ('Chief Engineer - S.O.', 'मुख्य अभियंता - एस. ओ.', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcesewerageoper', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlChiefEngSewOperations'),
  ('Chief Engineer - SWM', 'मुख्य अभियंता - मुख्य अभियंता (एस. डब्ल्यू. एम)', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlswmrtioverview', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlChiefEngSWM'),
  ('Chief Engineer - Storm Water Drains', 'प्रमुखअभियंता - पर्जन्य जल वाहिन्या', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcheswdpage', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcheswd'),
  ('Chief Engineer - Vigilance', 'प्रमुख अभिंयता - दक्षता', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcevigilance', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlchevigilance'),
  ('मुख्य निरीक्षक दुकाने व आस्थापना प्रमुख अधिकारी दुकाने व आस्थापना', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlChiefInspShopEst', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcise'),
  ('Chief Fire Officer', 'प्रमुख अग्निशमन अधिकारी', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrticfo', NULL),
  ('Chief Labour Officer', 'प्रमुख कामगार अधिकारी', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qllabourdept', NULL),
  ('Chief Officer - Enquiries', 'प्रमुख अधिकारी - चौकशी', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrtienquiries', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlchefenqu'),
  ('Civic Training Institute and Research Centre', 'नागरी प्रशिक्षण संस्था वसंशोधन केंद्र', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlctircenter', NULL),
  ('City Engineer', 'नगर अभियंता', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcityengineer', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcityengineerdocs'),
  ('Cooper Hospital', 'कूपर रूग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlcooperhosp', NULL),
  ('Chief Engineer - Water Supply Project', 'पाणी पुरवठाप्रकल्प', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlWaterSupplyProj', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlwsprodocs'),
  ('Deonar Abattoir', 'देवनार पशुवधगृह', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrtida', NULL),
  ('Deputy Chief Accountant - Establishment I', 'उप प्रमुख लेखापाल - आस्थापना १', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrtidcaest1', NULL),
  ('Deputy Chief Accountant - Establishment II', 'उप प्रमुख लेखापाल - आस्थापना २', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrtidcaest2', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlDeputyCAcc'),
  ('Deputy Chief Engineer (Building Proposal) Western Suburbs - II', 'उप प्रमुख अभियंता (ईमारत प्रस्ताव) पश्चिम उपनगरे - २', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qldcebpws', 'https://portal.mcgm.gov.in/irj/portal/anonymous/dycebpws2'),
  ('Deputy Chief Engineer - Privatization Initiative Cell', 'उप प्रमुख अभियंता - खाजगीकरण प्रारंभिक कक्ष', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qldcepvtncell', NULL),
  ('Deputy Municipal Commissioner - Special Engineering', 'उप आयुक्त - विशेष अभियांत्रिकी', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qldymcspecialeng', NULL),
  ('Director (Engineering Services & Projects)', 'संचालक (अभियांत्रिकी सेवा व प्रकल्प)', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qldesproj', NULL),
  ('Disaster Management Department', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qldmd', NULL),
  ('DMC (Vigilance)', 'उप आयुक्त (दक्षता)', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlDMCVig', NULL),
  ('DMC Infrastructure', 'उप आयुक्त पायाभूत सुविधा', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlDMCInfra'),
  ('DMC (Finance)', 'उप आयुक्त (वित्त)', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlDMCFinance'),
  ('Education', 'शिक्षण', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qledutnoff', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qleducationanon'),
  ('Environment Department', 'पर्यावरण विभागाच्या', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlEnvironment'),
  ('Estate Department', 'मालमत्ता विभाग', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlassicommestates', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlestate'),
  ('Gardens & Tree Authority', 'उद्याने व वृक्ष प्राधिकरण', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlgarden', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlgarden1'),
  ('Garden Cell', 'उद्यान कक्ष', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlgarden', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlgardencell'),
  ('Ghatkopar Swimming Pool', 'घाटकोपर जलतरण तलाव', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlGhatkoparSwimmingPool', NULL),
  ('Seth G.S.Medical College & KEM Hospital', 'सेठ गो.सु.वैद्यकीय महाविद्यालय व रा.ए.स्मारक रुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlgskemhosprti', NULL),
  ('Hydraulic Engineer', 'जल अभियंता', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlhedept', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlhedocs'),
  ('Information Technology', 'माहितीतंत्रज्ञान', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlitdept1', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlITanonymous'),
  ('Insecticide Department', 'कीटकनाशकविभाग', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlinsecticide', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlisd'),
  ('Kasturba Hospital', 'कस्तुरबा रुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlkasturbahosp', NULL),
  ('K.B.Bhabha Municipal General Hospital, Bandra', 'के.बी.भाभाम.न.पा.सर्वसाधारण रुग्णालय, बांद्रा', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlkbbmunicipalhosp', NULL),
  ('K.B.Bhabha Municipal General Hospital, Kurla', 'के.बी.भाभाम.न.पा.सर्वसाधारण रुग्णालय, कुर्ला', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlkbbhabamunhosp', NULL),
  ('K.M.J.Phule Municipal General Hospital', 'क्रांतीवीर महात्मा ज्योतिबा फुले महापालिका सर्वसाधारणरुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlkmjphulehosp', NULL),
  ('Legal Department', 'विधि खाते', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qllegal', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qllegaldep'),
  ('Licence', 'अनुज्ञापन', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qllicdept', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qllicdocs'),
  ('Maa Hospital', 'माँ रुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlMaaHospital', NULL),
  ('Manager, Prabodhankar K.S. Thackarey Natyamandir, Borivali', 'व्यवस्थापक, प्रबोधनकार ठाकरेनाट्यमंदीर, बोरिवली', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmanagerptn', NULL),
  ('Master Deenanath Mangeshkar Natyagriha', 'मास्टर दीनानाथ मंगेशकर नाट्यग्रह', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qldinanath', NULL),
  ('Market Department', 'बाजार विभाग', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrtimrktdept', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmdep'),
  ('Municipal Commissioner''s Office', 'महापालिका आयुक्त कार्यालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmcoffice', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmunoffchief'),
  ('मुंबई मलनि:सारण प्रकल्प Mumbai Sewerage Disposal Project', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmumbaisewdisproj', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmsdp'),
  ('Mumbai Fire Brigade', 'मुंबई अग्निशमन दल', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrticfo', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlMumFireBrigadeAnon'),
  ('Municipal Chief Auditor', 'मुख्यलेखा परीक्षक', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmcauditors', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmcauditorsanonymous'),
  ('Municipal Eye Hospital', 'महापालिकानेत्र रुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmunieyehosp', NULL),
  ('Municipal Secretary', 'महानगरपालिका चिटणीस', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlMunicipalSecretary', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlmsanonymous'),
  ('Municipal Analyst Laboratory', 'महापालिका विश्लेषक,प्रयोगशाळा', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlMunicipalAnalystLaboratory', NULL),
  ('मुरली दे वरा मुनपा नेत्र रुग्णालीय Murli Deora Municipal Eye Hospital', NULL, 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlMurliDeoraEyeHosp', NULL),
  ('Nair Hospital Dental College', 'नायर रूग्णालयदंत महाविद्यालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlnhdc', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlnhdcd'),
  ('Pandit Madanmohan Malviya Shatabdi Hospital, Govandi', 'शताब्दी रूग्णालय, गोवंडी', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlShatabdiHosp', NULL),
  ('Prabodhankar Thakare Open Theatre, Sewri', 'प्रबोधनकार ठाकरे खुले नाट्यगृह, शिवडी', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlptots', NULL),
  ('Public Health', 'सार्वजनिक आरोग्य', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlpublichealthdept', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlpublichealthdeptAnon'),
  ('Public Relation', 'जनसंपर्क', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrtipro', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlprodocs'),
  ('Raja Baldevdas Birla Kreeda Kendra', 'राजाबलदेवदास बिर्ला क्रिडा केन्द्र', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlbkk', NULL),
  ('Roads & Traffic', 'रस्ते व वाहतुक', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlroadsandtraffic', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlroadsdocs'),
  ('Sardar Vallabhbhai Patel Swimming Pool', 'सरदार वल्लभभाई पटेल तरणतलाव', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlsvpatelspool', NULL),
  ('Security', 'सुरक्षा', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlrticsecdep', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlsecdepanonymous'),
  ('S.K. Patil Municipal General Hospital', 'स.का.पाटील महापालिका सर्वसाधारण रुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlskpmunhosp', NULL),
  ('Seth A.J.B. Municipal E.N.T.hospital', 'सेठ ए.जे.बी. महापालिका कान, नाक, घसा रुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlenthosp', NULL),
  ('Shatabdi Hospital, Kandivli', 'शताब्दी रूग्णालय, कांदिवली', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlshatabdihospkandivli', NULL),
  ('Test Audit & Vigilance Officer', 'चाचणी लेखा परिक्षा व दक्षता अधिकारी', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qltestvigillanceoff', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qltavo'),
  ('Vigilance Officer - Special Duty', 'दक्षता अधिकारी - विशेषकार्य', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlvigoffspduty', NULL),
  ('V.D.Savarkar General Hospital, Mulund', 'वि.दा.सावरकार सर्वसाधारण रूग्णालय, मुलूंड', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlvsavarkarhosp', NULL),
  ('V.N.Desai Municipal General Hospital', 'वि.न.देसाई महापालिका सर्वसाधारणरुग्णालय', 'https://portal.mcgm.gov.in/irj/portal/anonymous/qlvndmunhosp', NULL)
ON CONFLICT (name_en) DO NOTHING;
